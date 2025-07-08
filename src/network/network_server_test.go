/******************************************************************************/
/* network_server_test.go                                                     */
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
	"fmt"
	"testing"
)

func TestFlushPending1(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}
	orders := []uint64{1, 3, 4, 0}
	lens := []int{1, 2, 3, 2}
	msg := []byte("test")
	p := NetworkPacketUDP{
		order:      1,
		messageLen: uint16(len(msg)),
	}
	copy(p.message[:], msg)
	for i := range orders {
		p.order = orders[i]
		c.flushPending(p, &s.ClientMessageQueue)
		if len(c.reliableBuffer) != lens[i] {
			t.FailNow()
		}
	}
}

func TestFlushPending2(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}
	orders := []uint64{1, 3, 4, 2, 0}
	lens := []int{1, 2, 3, 4, 0}
	msg := []byte("test")
	p := NetworkPacketUDP{
		order:      1,
		messageLen: uint16(len(msg)),
	}
	copy(p.message[:], msg)
	for i := range orders {
		p.order = orders[i]
		c.flushPending(p, &s.ClientMessageQueue)
		if len(c.reliableBuffer) != lens[i] {
			t.FailNow()
		}
	}
}

func TestReliableMessageQueue(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}
	orders := []uint64{1, 3, 4, 2, 0}
	lens := []int{1, 2, 3, 4, 0}
	for i := range orders {
		msg := []byte(fmt.Sprintf("Test: %d", orders[i]))
		p := NetworkPacketUDP{
			order:      orders[i],
			messageLen: uint16(len(msg)),
		}
		copy(p.message[:], msg)
		c.flushPending(p, &s.ClientMessageQueue)
		if len(c.reliableBuffer) != lens[i] {
			t.FailNow()
		}
	}
	messages := s.ClientMessageQueue.Flush()
	if len(messages) != len(orders) {
		t.FailNow()
	}
	for i := range orders {
		if string(messages[i].Message()) != fmt.Sprintf("Test: %d", i) {
			t.FailNow()
		}
	}
}
