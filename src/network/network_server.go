/******************************************************************************/
/* network_server.go                                                          */
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

package network

import (
	"fmt"
	"kaiju/engine"
	"kaiju/platform/concurrent"
	"log/slog"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type ClientMessage struct {
	message    [maxPacketSize]byte
	messageLen uint16
	Client     *ServerClient
}

func (c *ClientMessage) IsFromServer() bool {
	return c.Client == nil || c.Client.addr == nil
}

func clientMessageFromPacket(packet NetworkPacketUDP, client *ServerClient) ClientMessage {
	cm := ClientMessage{
		Client:     client,
		messageLen: packet.messageLen,
	}
	copy(cm.message[:], packet.message[:packet.messageLen])
	return cm
}

func (c *ClientMessage) Message() []byte { return c.message[:c.messageLen] }

type ServerClient struct {
	id             int
	addr           *net.UDPAddr
	writeBuffer    []byte
	readBuffer     []byte
	reliableBuffer []NetworkPacketUDP
	reliableOrder  uint64
	writeMutex     sync.Mutex
}

type NetworkServer struct {
	NetworkUDP
	ClientMessageQueue concurrent.MessageQueue[ClientMessage]
	clients            map[string]*ServerClient
	nextClientId       int
}

func NewServerUDP() NetworkServer {
	return NetworkServer{
		clients: make(map[string]*ServerClient),
	}
}

func (c *ServerClient) Id() int         { return c.id }
func (c *ServerClient) Address() string { return c.addr.String() }

func (c *ServerClient) PortlessAddress() string {
	full := c.addr.String()
	addr := strings.Index(full, ":")
	return full[:addr]
}

func (s *NetworkServer) addClient(addr *net.UDPAddr) *ServerClient {
	client := &ServerClient{
		id:          s.nextClientId,
		addr:        addr,
		writeBuffer: make([]byte, maxPacketSize),
		readBuffer:  make([]byte, maxPacketSize),
	}
	s.clients[addr.String()] = client
	s.nextClientId++
	return client
}

func (s *NetworkServer) RemoveClient(client *ServerClient) {
	delete(s.clients, client.addr.String())
}

func (s *NetworkServer) HolePunchClient(address string, port uint16) (*ServerClient, error) {
	portStr := strconv.Itoa(int(port))
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(address, portStr))
	if err != nil {
		slog.Error("failed to resolve the client address for a hole punch", "address", address, "port", port)
	}
	return &ServerClient{
		id:          0,
		addr:        addr,
		writeBuffer: make([]byte, maxPacketSize),
		readBuffer:  make([]byte, maxPacketSize),
	}, err
}

func (s *NetworkServer) Serve(updater *engine.Updater, port uint16) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("failed to resolve the UDP host address", "error", err, "port", port)
		return err
	}
	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		slog.Error("failed to dial the UDP server", "error", err, "port", port)
		return err
	}
	slog.Info("UDP server started listening", "port", port)
	s.updateId = updater.AddUpdate(s.update)
	go s.readMessages()
	return nil
}

func (s *NetworkServer) sendPacket(packet NetworkPacketUDP, client *ServerClient) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()
	n, err := packetToMessage(packet, client.writeBuffer)
	if err != nil {
		return err
	}
	_, err = s.conn.WriteToUDP(client.writeBuffer[:n], client.addr)
	if err != nil {
		slog.Error("failed to write message to client", "error", err, "client", client)
		return err
	}
	return nil
}

func (c *NetworkServer) SendMessageUnreliable(message []byte, client *ServerClient) error {
	return c.sendPacket(c.createUnreliable(message), client)
}

func (c *NetworkServer) SendMessageReliable(message []byte, client *ServerClient) error {
	return c.sendPacket(c.createReliable(message, client), client)
}

func (s *NetworkServer) readMessages() {
	s.isReading = true
	readBuffer := make([]byte, maxPacketSize)
	for s.isReading {
		n, remoteAddr, err := s.conn.ReadFromUDP(readBuffer)
		if !s.isReading {
			break
		}
		if err != nil {
			slog.Error("failed reading client message", "error", err)
			continue
		}
		client, ok := s.clients[remoteAddr.String()]
		if !ok {
			client = s.addClient(remoteAddr)
		}
		copy(client.readBuffer, readBuffer)
		packet := packetFromMessage(client.readBuffer[:n])
		if packet.isAck() {
			s.removePendingPacket(packet.timestamp)
		} else {
			if packet.isReliable() {
				// The ack is just the timestamp of the message it read
				s.sendPacket(s.createAck(client.readBuffer[:unsafe.Sizeof(packet.timestamp)]), client)
				client.flushPending(packet, &s.ClientMessageQueue)
			} else {
				s.ClientMessageQueue.Enqueue(clientMessageFromPacket(packet, client))
			}
		}
	}
	slog.Info("UDP network server stopped reading messages")
}

func (client *ServerClient) flushPending(p NetworkPacketUDP, messageQueue *concurrent.MessageQueue[ClientMessage]) {
	if p.order < client.reliableOrder {
		// We already have processed this packet
		return
	}
	if p.order == client.reliableOrder {
		client.reliableBuffer = append(client.reliableBuffer, p)
		// Reverse the list so that the lowest id (the one we're on) is at the end
		sort.Slice(client.reliableBuffer, func(i, j int) bool {
			return client.reliableBuffer[i].order > client.reliableBuffer[j].order
		})
		// Go backwards through the list until we hit an id we're not ready for
		end := len(client.reliableBuffer) - 1
		for ; end >= 0; end-- {
			if client.reliableBuffer[end].order == client.reliableOrder {
				messageQueue.Enqueue(clientMessageFromPacket(client.reliableBuffer[end], client))
				// Go to the next reliable message id
				client.reliableOrder++
			} else {
				break
			}
		}
		// Remove all of the processed messages from the end
		client.reliableBuffer = client.reliableBuffer[:end+1]
	} else {
		for i := range client.reliableBuffer {
			if p.order == client.reliableBuffer[i].order {
				// We've already added this reliable packet to the list
				return
			}
		}
		client.reliableBuffer = append(client.reliableBuffer, p.clone())
	}
}

func (s *NetworkServer) update(deltaTime float64) {
	if len(s.pendingPackets) == 0 {
		return
	}
	now := time.Now()
	s.pendingMutex.RLock()
	defer s.pendingMutex.RUnlock()
	for i := 0; i < len(s.pendingPackets); i++ {
		pp := &s.pendingPackets[i]
		if pp.packet.nextRetry.Before(now) {
			pp.packet.nextRetry = now.Add(reliableRetryDelay)
			s.sendPacket(pp.packet, pp.target)
		}
	}
}
