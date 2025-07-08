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
