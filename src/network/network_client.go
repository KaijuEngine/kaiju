/******************************************************************************/
/* network_client.go                                                          */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
	"encoding/binary"
	"fmt"
	"kaiju/engine"
	"kaiju/platform/concurrent"
	"log/slog"
	"net"
	"time"
	"unsafe"
)

type NetworkClient struct {
	NetworkUDP
	ServerClient
	ServerMessageQueue concurrent.MessageQueue[ClientMessage]
}

func NewClientUDP() NetworkClient {
	return NetworkClient{
		ServerClient: ServerClient{
			readBuffer:  make([]byte, maxPacketSize),
			writeBuffer: make([]byte, maxPacketSize),
		},
	}
}

func (c *NetworkClient) Connect(updater *engine.Updater, address string, port uint16) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		slog.Error("failed to resolve the UDP host address", "error", err, "address", address, "port", port)
		return err
	}
	c.conn, err = net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		slog.Error("failed to dial the UDP server", "error", err, "address", address, "port", port)
		return err
	}
	c.updateId = updater.AddUpdate(c.update)
	go c.ReadMessages()
	return nil
}

func (c *NetworkClient) sendPacket(packet NetworkPacketUDP) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	n, err := packetToMessage(packet, c.writeBuffer)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(c.writeBuffer[:n])
	if err != nil {
		slog.Error("error writing message from client to server", "error", err, "packet", packet)
		return err
	}
	return nil
}

func (c *NetworkClient) SendMessageUnreliable(message []byte) error {
	return c.sendPacket(c.createUnreliable(message))
}

func (c *NetworkClient) SendMessageReliable(message []byte) error {
	return c.sendPacket(c.createReliable(message, &c.ServerClient))
}

func (c *NetworkClient) ReadMessages() {
	slog.Info("UDP network client starting message read pipeline")
	c.isReading = true
	buffer := make([]byte, maxPacketSize)
	for c.isReading {
		//c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := c.conn.Read(buffer)
		if !c.isReading {
			break
		}
		if err != nil {
			slog.Error("the UDP network client failed to read", "error", err)
			c.isReading = false
			break
		}
		packet := packetFromMessage(buffer[:n])
		if packet.isAck() {
			if uintptr(packet.messageLen) == unsafe.Sizeof(packet.timestamp) {
				id := int64(binary.LittleEndian.Uint64(packet.message[:]))
				c.removePendingPacket(id)
			}
		} else {
			if packet.isReliable() {
				// The ack is just the timestamp of the message it read
				c.sendPacket(c.createAck(buffer[:unsafe.Sizeof(packet.timestamp)]))
				if packet.order >= c.reliableOrder {
					c.flushPending(packet, &c.ServerMessageQueue)
				}
			} else {
				c.ServerMessageQueue.Enqueue(clientMessageFromPacket(packet, nil))
			}
		}
	}
	slog.Info("UDP network client stopped reading messages")
}

func (s *NetworkClient) update(deltaTime float64) {
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
			s.sendPacket(pp.packet)
		}
	}
}
