/******************************************************************************/
/* master_server.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package master_server

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/network"
	"log/slog"
	"maps"
	"time"
	"unsafe"
)

const (
	MaxPasswordSize = 16

	masterAddress = "localhost"
	masterPort    = 15973
	serverTimeout = time.Second * 30
	gameKeySize   = 32
	gameNameSize  = 64
)

type MasterServer struct {
	server     network.NetworkServer
	serverList map[int]ServerListing
}

type ServerListing struct {
	game           string
	name           string
	password       string
	client         *network.ServerClient
	maxPlayers     uint16
	currentPlayers uint16
	timeoutAt      time.Time
}

func New(updater *engine.Updater) (*MasterServer, error) {
	ms := &MasterServer{
		server:     network.NewServerUDP(),
		serverList: make(map[int]ServerListing),
	}
	err := ms.server.Serve(updater, masterPort)
	updater.AddUpdate(ms.update)
	return ms, err
}

func (m *MasterServer) update(float64) {
	messages := m.server.ClientMessageQueue.Flush()
	for i := range messages {
		m.processMessage(messages[i])
	}
	m.evictUnresponsiveServers()
}

func (m *MasterServer) evictUnresponsiveServers() {
	now := time.Now()
	keys := maps.Keys(m.serverList)
	for k := range keys {
		serv := m.serverList[k]
		if serv.timeoutAt.Before(now) {
			slog.Info("Game server has timed out", "address", serv.client.Address())
			m.server.RemoveClient(serv.client)
			delete(m.serverList, k)
		}
	}
}

func (m *MasterServer) processMessage(msg network.ClientMessage) {
	buffer := msg.Message()
	if len(buffer) != int(unsafe.Sizeof(Request{})) {
		return
	}
	req := DeserializeRequest(buffer)
	if _, exists := m.serverList[msg.Client.Id()]; exists {
		m.processClientMessage(req, msg)
	} else if req.Type == RequestTypeRegisterServer {
		debug.Log("<- Register")
		m.processNewServer(req, msg)
	} else {
		m.processClientRequestMessage(req, msg)
	}
}

func (m *MasterServer) processClientMessage(req Request, msg network.ClientMessage) {
	id := msg.Client.Id()
	switch req.Type {
	case RequestTypeUnregisterServer:
		debug.Log("<- Unregister")
		delete(m.serverList, id)
	case RequestTypeRegisterServer:
		fallthrough
	case RequestTypePing:
		debug.Log("<- Ping")
		serv := m.serverList[id]
		serv.currentPlayers = req.CurrentPlayers
		serv.timeoutAt = time.Now().Add(serverTimeout)
		m.serverList[id] = serv
	}
}

func (m *MasterServer) processNewServer(req Request, msg network.ClientMessage) {
	if req.Type != RequestTypeRegisterServer {
		return
	}
	err := m.sendResponse(Response{Type: ResponseTypeConfirmRegister}, msg.Client)
	if err == nil {
		listing := ServerListing{
			game:       klib.ByteArrayToString(req.Game[:]),
			name:       klib.ByteArrayToString(req.Name[:]),
			password:   klib.ByteArrayToString(req.Password[:]),
			client:     msg.Client,
			maxPlayers: req.MaxPlayers,
			timeoutAt:  time.Now().Add(serverTimeout),
		}
		m.serverList[msg.Client.Id()] = listing
	}
}

func (m *MasterServer) processClientRequestMessage(req Request, msg network.ClientMessage) {
	switch req.Type {
	case RequestTypeServerList:
		debug.Log("<- Server list")
		m.sendServerList(req, msg)
	case RequestTypeJoinServer:
		debug.Log("<- Join server")
		if serv, ok := m.serverList[int(req.ServerId)]; ok {
			if serv.password == klib.ByteArrayToString(req.Password[:]) {
				res := Response{Type: ResponseTypeJoinServerInfo}
				copy(res.Address[:], serv.client.PortlessAddress())
				m.sendResponse(res, msg.Client)
				servRes := Response{Type: ResponseTypeClientJoinInfo}
				copy(servRes.Address[:], msg.Client.PortlessAddress())
				m.sendResponse(servRes, serv.client)
			} else {
				m.sendResponse(Response{Type: ResponseTypeError, Error: ErrorIncorrectPassword}, msg.Client)
			}
		} else {
			m.sendResponse(Response{Type: ResponseTypeError, Error: ErrorServerDoesntExist}, msg.Client)
		}
	}
}

func (m *MasterServer) sendServerList(req Request, msg network.ClientMessage) {
	count := 0
	var res Response
	game := klib.ByteArrayToString(req.Game[:])
	totalCount := uint32(0)
	for k := range m.serverList {
		if m.serverList[k].game == game {
			totalCount++
		}
	}
	for k := range m.serverList {
		if count == 0 {
			res = Response{
				Type:      ResponseTypeServerList,
				TotalList: totalCount,
			}
		}
		serv := m.serverList[k]
		if serv.game != game {
			continue
		}
		res.List[count] = ResponseServerList{
			Id:             uint64(serv.client.Id()),
			MaxPlayers:     serv.maxPlayers,
			CurrentPlayers: serv.currentPlayers,
		}
		copy(res.List[count].Name[:], serv.name)
		count++
		if count == len(res.List) {
			m.sendResponse(res, msg.Client)
			count = 0
		}
	}
	if count != 0 {
		m.sendResponse(res, msg.Client)
	}
}

func (m *MasterServer) sendResponse(res Response, client *network.ServerClient) error {
	switch res.Type {
	case ResponseTypeConfirmRegister:
		debug.Log("-> Confirm register")
	case ResponseTypeServerList:
		debug.Log("-> Server list")
	case ResponseTypeJoinServerInfo:
		debug.Log("-> Join server info")
	case ResponseTypeClientJoinInfo:
		debug.Log("-> Client join info")
	case ResponseTypeError:
		debug.Log("-> Error", "error", res.Error)
	}
	buff := [unsafe.Sizeof(Response{})]byte{}
	res.Serialize(buff[:])
	return m.server.SendMessageReliable(buff[:], client)
}
