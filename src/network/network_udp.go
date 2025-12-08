/******************************************************************************/
/* network_udp.go                                                             */
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
	"net"
	"sync"
	"time"

	"github.com/KaijuEngine/kaiju/engine"
	"github.com/KaijuEngine/kaiju/klib"
)

const reliableRetryDelay = time.Millisecond * 15

type PendingNetworkPacketUDP struct {
	target *ServerClient
	packet NetworkPacketUDP
}

type NetworkUDP struct {
	conn           *net.UDPConn
	pendingPackets []PendingNetworkPacketUDP
	pendingMutex   sync.RWMutex
	updateId       engine.UpdateId
	isReading      bool
}

func (n *NetworkUDP) IsLive() bool { return n.conn != nil }

func (n *NetworkUDP) Close(updater *engine.Updater) {
	n.isReading = false
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	updater.RemoveUpdate(&n.updateId)
}

func (n *NetworkUDP) removePendingPacket(id int64) {
	n.pendingMutex.Lock()
	defer n.pendingMutex.Unlock()
	for i := range n.pendingPackets {
		if n.pendingPackets[i].packet.timestamp == id {
			n.pendingPackets = klib.RemoveUnordered(n.pendingPackets, i)
			break
		}
	}
}

func (n *NetworkUDP) createUnreliable(message []byte) NetworkPacketUDP {
	packet := NetworkPacketUDP{
		timestamp:  time.Now().UTC().UnixMicro(),
		messageLen: uint16(len(message)),
	}
	copy(packet.message[:], message)
	return packet
}

func (n *NetworkUDP) createReliable(message []byte, target *ServerClient) NetworkPacketUDP {
	packet := NetworkPacketUDP{
		timestamp:  time.Now().UTC().UnixMicro(),
		order:      target.reliableOrder,
		messageLen: uint16(len(message)),
		typeFlags:  udpPacketTypeReliable,
		nextRetry:  time.Now().Add(reliableRetryDelay),
	}
	target.reliableOrder++
	copy(packet.message[:], message)
	n.pendingMutex.Lock()
	n.pendingPackets = append(n.pendingPackets, PendingNetworkPacketUDP{
		target: target,
		packet: packet,
	})
	n.pendingMutex.Unlock()
	return packet
}

func (n *NetworkUDP) createAck(fromTimestamp []byte) NetworkPacketUDP {
	packet := NetworkPacketUDP{
		timestamp:  time.Now().UTC().UnixMicro(),
		messageLen: uint16(len(fromTimestamp)),
		typeFlags:  udpPacketTypeAck,
	}
	copy(packet.message[:], fromTimestamp)
	return packet
}
