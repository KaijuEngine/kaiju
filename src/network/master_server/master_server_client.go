/******************************************************************************/
/* master_server_client.go                                                    */
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
	"unsafe"
)

const (
	pingIntervalSeconds = 3.0
)

type MasterServerClient struct {
	client       network.NetworkClient
	pingTime     float64
	updateId     engine.UpdateId
	isServer     bool
	OnServerList func([]ResponseServerList, uint32)
	OnServerJoin func(string)
	OnClientJoin func(string)
	OnError      func(uint8)
}

func (c *MasterServerClient) Connect(updater *engine.Updater) error {
	if c.updateId != 0 {
		c.Disconnect(updater)
	}
	c.client = network.NewClientUDP()
	if c.OnServerList == nil {
		c.OnServerList = func([]ResponseServerList, uint32) {}
	}
	if c.OnServerJoin == nil {
		c.OnServerJoin = func(s string) {}
	}
	if c.OnClientJoin == nil {
		c.OnClientJoin = func(s string) {}
	}
	if c.OnError == nil {
		c.OnError = func(u uint8) {}
	}
	err := c.client.Connect(updater, masterAddress, masterPort)
	if err != nil {
		slog.Error("failed to setup connection for master server", "address", masterAddress, "port", masterPort)
	} else {
		debug.Log("Successfully bound master server client")
	}
	c.updateId = updater.AddUpdate(c.update)
	return err
}

func (c *MasterServerClient) Disconnect(updater *engine.Updater) {
	debug.Log("Disconnecting the master server client")
	updater.RemoveUpdate(&c.updateId)
	c.client.Close(updater)
}

func (c *MasterServerClient) RegisterServer(game, name string, maxPlayers, currentPlayers uint16) error {
	debug.Log("Registering server with master server")
	req := Request{
		Type:           RequestTypeRegisterServer,
		MaxPlayers:     maxPlayers,
		CurrentPlayers: currentPlayers,
	}
	copy(req.Game[:], game)
	copy(req.Name[:], name)
	c.isServer = true
	return c.sendRequest(req)
}

func (c *MasterServerClient) ListServers(game string) error {
	req := Request{Type: RequestTypeServerList}
	copy(req.Game[:], game)
	return c.sendRequest(req)
}

func (c *MasterServerClient) JoinServer(id uint64) error {
	return c.sendRequest(Request{Type: RequestTypeJoinServer, ServerId: id})
}

func (c *MasterServerClient) update(deltaTime float64) {
	if c.isServer {
		c.pingTime -= deltaTime
		if c.pingTime < 0 {
			c.pingTime = pingIntervalSeconds
			c.sendRequest(Request{Type: RequestTypePing})
		}
	}
	messages := c.client.ServerMessageQueue.Flush()
	for i := range messages {
		buff := messages[i].Message()
		if len(buff) != int(unsafe.Sizeof(Response{})) {
			continue
		}
		c.processMessage(messages[i])
	}
}

func (c *MasterServerClient) processMessage(msg network.ClientMessage) {
	res := DeserializeResponse(msg.Message())
	switch res.Type {
	case ResponseTypeConfirmRegister:
		debug.Log("<- Confirm register")
	case ResponseTypeServerList:
		debug.Log("<- Server list")
		c.OnServerList(res.List[:], res.TotalList)
	case ResponseTypeJoinServerInfo:
		debug.Log("<- Join server")
		c.OnServerJoin(klib.ByteArrayToString(res.Address[:]))
	case ResponseTypeClientJoinInfo:
		debug.Log("<- Client join")
		c.OnClientJoin(klib.ByteArrayToString(res.Address[:]))
	case ResponseTypeError:
		debug.Log("<- Error", "error", res.Error)
		c.OnError(res.Error)
	}
}

func (c *MasterServerClient) sendRequest(req Request) error {
	switch req.Type {
	case RequestTypeRegisterServer:
		debug.Log("-> Register")
	case RequestTypeUnregisterServer:
		debug.Log("-> Unregister")
	case RequestTypePing:
		debug.Log("-> Ping")
	case RequestTypeServerList:
		debug.Log("-> Server list")
	case RequestTypeJoinServer:
		debug.Log("-> Connect to server")
	}
	buff := [unsafe.Sizeof(Request{})]byte{}
	req.Serialize(buff[:])
	err := c.client.SendMessageReliable(buff[:])
	if err != nil {
		slog.Error("failed to send the message to the master server")
	}
	return err
}
