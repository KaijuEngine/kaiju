/******************************************************************************/
/* network_packet_udp.go                                                      */
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
	"encoding/binary"
	"errors"
	"log/slog"
	"time"
	"unsafe"
)

const maxPacketSize = 1024

type udpPacketTypeFlags = uint32

const (
	udpPacketTypeReliable = udpPacketTypeFlags(1 << 0)
	udpPacketTypeAck      = udpPacketTypeFlags(1 << 1)
)

type NetworkPacketUDP struct {
	timestamp  int64
	order      uint64
	messageLen uint16
	message    [maxPacketSize]byte
	typeFlags  udpPacketTypeFlags
	nextRetry  time.Time
}

func (p *NetworkPacketUDP) isReliable() bool {
	return p.typeFlags&udpPacketTypeReliable != 0
}

func (p *NetworkPacketUDP) isAck() bool {
	return p.typeFlags&udpPacketTypeAck != 0
}

func (p *NetworkPacketUDP) clone() NetworkPacketUDP {
	c := NetworkPacketUDP{
		timestamp:  p.timestamp,
		order:      p.order,
		messageLen: p.messageLen,
		typeFlags:  p.typeFlags,
		nextRetry:  p.nextRetry,
	}
	copy(c.message[:], p.message[:])
	return c
}

func packetToMessage(packet NetworkPacketUDP, buffer []byte) (int, error) {
	totalSize := unsafe.Sizeof(packet.timestamp) +
		unsafe.Sizeof(packet.order) +
		unsafe.Sizeof(packet.messageLen) +
		uintptr(packet.messageLen) +
		unsafe.Sizeof(packet.typeFlags)
	if uintptr(len(buffer)) < totalSize {
		const err = "buffer is not large enough to hold the message"
		slog.Error(err, "bufferSize", len(buffer), "dataSize", totalSize)
		return 0, errors.New(err)
	}
	p := uintptr(0)
	binary.LittleEndian.PutUint64(buffer[p:], uint64(packet.timestamp))
	p += unsafe.Sizeof(packet.timestamp)
	binary.LittleEndian.PutUint64(buffer[p:], packet.order)
	p += unsafe.Sizeof(packet.order)
	binary.LittleEndian.PutUint16(buffer[p:], packet.messageLen)
	p += unsafe.Sizeof(packet.messageLen)
	p += uintptr(copy(buffer[p:], packet.message[:packet.messageLen]))
	binary.LittleEndian.PutUint32(buffer[p:], packet.typeFlags)
	return int(totalSize), nil
}

func packetFromMessage(message []byte) NetworkPacketUDP {
	packet := NetworkPacketUDP{}
	p := uintptr(0)
	packet.timestamp = int64(binary.LittleEndian.Uint64(message[p:]))
	p += unsafe.Sizeof(packet.timestamp)
	packet.order = binary.LittleEndian.Uint64(message[p:])
	p += unsafe.Sizeof(packet.order)
	packet.messageLen = binary.LittleEndian.Uint16(message[p:])
	p += unsafe.Sizeof(packet.messageLen)
	p += uintptr(copy(packet.message[:], message[p:p+uintptr(packet.messageLen)]))
	packet.typeFlags = binary.LittleEndian.Uint32(message[p:])
	return packet
}
