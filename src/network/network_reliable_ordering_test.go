/******************************************************************************/
/* network_reliable_ordering_test.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package network

import (
	"testing"
)

func createTestPacket(order uint64, msg string) NetworkPacketUDP {
	p := NetworkPacketUDP{
		order:      order,
		messageLen: uint16(len(msg)),
	}
	copy(p.message[:], []byte(msg))
	return p
}

func TestFlushPending_SequentialArrival(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	for i := uint64(0); i < 5; i++ {
		p := createTestPacket(i, string(rune('A'+i)))
		c.flushPending(p, &s.ClientMessageQueue)
	}

	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer after sequential flush, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 5 {
		t.Errorf("reliableOrder = %d, want 5", c.reliableOrder)
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(msgs))
	}
	for i, m := range msgs {
		expected := string(rune('A' + i))
		if string(m.Message()) != expected {
			t.Errorf("msg[%d] = %q, want %q", i, string(m.Message()), expected)
		}
	}
}

func TestFlushPending_GapThenFill(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send 0, then 3 (gap), then 1, then 2
	orders := []uint64{0, 3, 1, 2}
	for _, o := range orders {
		p := createTestPacket(o, string(rune('A'+o)))
		c.flushPending(p, &s.ClientMessageQueue)
	}

	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer after gap-fill, got %d", len(c.reliableBuffer))
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	for i, m := range msgs {
		expected := string(rune('A' + i))
		if string(m.Message()) != expected {
			t.Errorf("msg[%d] = %q, want %q", i, string(m.Message()), expected)
		}
	}
}

func TestFlushPending_PartialGap(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send 0, then 3 (gap for 1,2)
	c.flushPending(createTestPacket(0, "A"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(3, "D"), &s.ClientMessageQueue)

	if len(c.reliableBuffer) != 1 {
		t.Errorf("expected 1 buffered packet, got %d", len(c.reliableBuffer))
	}
	if c.reliableBuffer[0].order != 3 {
		t.Errorf("buffered order = %d, want 3", c.reliableBuffer[0].order)
	}

	// Fill the gap with 1 and 2
	c.flushPending(createTestPacket(1, "B"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(2, "C"), &s.ClientMessageQueue)

	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer after filling gap, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 4 {
		t.Errorf("reliableOrder = %d, want 4", c.reliableOrder)
	}
}

func TestFlushPending_DuplicatePacket(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	p := createTestPacket(0, "A")
	c.flushPending(p, &s.ClientMessageQueue)
	c.flushPending(p, &s.ClientMessageQueue)

	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer, got %d", len(c.reliableBuffer))
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 1 {
		t.Errorf("expected 1 message (not duplicated), got %d", len(msgs))
	}
}

func TestFlushPending_DuplicateOutOfOrderPacket(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	p1 := createTestPacket(1, "B")
	p2 := createTestPacket(3, "D")

	// Send 1 (buffered), then 3 (buffered)
	c.flushPending(p1, &s.ClientMessageQueue)
	c.flushPending(p2, &s.ClientMessageQueue)

	initialLen := len(c.reliableBuffer)
	if initialLen != 2 {
		t.Fatalf("expected 2 buffered after initial sends, got %d", initialLen)
	}

	// Duplicate 1
	c.flushPending(p1, &s.ClientMessageQueue)

	if len(c.reliableBuffer) != initialLen {
		t.Errorf("buffer changed after duplicate: was %d, now %d", initialLen, len(c.reliableBuffer))
	}
}

func TestFlushPending_FuturePacketsOnly(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send packets 5, 6, 7 when reliableOrder is 0
	for i := uint64(5); i <= 7; i++ {
		c.flushPending(createTestPacket(i, string(rune('A'+i))), &s.ClientMessageQueue)
	}

	// None should be dequeued since 0 is missing
	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 0 {
		t.Errorf("expected 0 flushed messages, got %d", len(msgs))
	}
	if len(c.reliableBuffer) != 3 {
		t.Errorf("expected 3 buffered packets, got %d", len(c.reliableBuffer))
	}
}

func TestFlushPending_StartingFromHigherOrder(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 5}

	// Send 4 (already processed), should be skipped
	c.flushPending(createTestPacket(4, "old"), &s.ClientMessageQueue)
	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected 0 buffer (packet 4 already processed), got %d", len(c.reliableBuffer))
	}

	// Send 5 (expected next)
	c.flushPending(createTestPacket(5, "E"), &s.ClientMessageQueue)
	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected 0 buffer after processing 5, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 6 {
		t.Errorf("reliableOrder = %d, want 6", c.reliableOrder)
	}
}

func TestFlushPending_LargeGap(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send 0, then jump to 20
	c.flushPending(createTestPacket(0, "A"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(20, "Z"), &s.ClientMessageQueue)

	if len(c.reliableBuffer) != 1 {
		t.Errorf("expected 1 buffered packet (20), got %d", len(c.reliableBuffer))
	}

	// Fill the gap from 1 to 19
	for i := uint64(1); i < 20; i++ {
		c.flushPending(createTestPacket(i, string(rune('A'+(i%26)))), &s.ClientMessageQueue)
	}

	// Now 20 should also be processed
	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer after filling gap, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 21 {
		t.Errorf("reliableOrder = %d, want 21", c.reliableOrder)
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 21 {
		t.Errorf("expected 21 messages, got %d", len(msgs))
	}
}

func TestFlushPending_OrderPreservedInQueue(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send in order 3, 2, 1, 0
	orders := []uint64{3, 2, 1, 0}
	for _, o := range orders {
		c.flushPending(createTestPacket(o, string(rune('A'+o))), &s.ClientMessageQueue)
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	for i, m := range msgs {
		expected := string(rune('A' + i))
		if string(m.Message()) != expected {
			t.Errorf("msg[%d] = %q, want %q (order not preserved)", i, string(m.Message()), expected)
		}
	}
}

func TestFlushPending_AltersReliableOrderOnlyOnFlush(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send 0 - should be flushed, reliableOrder becomes 1
	c.flushPending(createTestPacket(0, "A"), &s.ClientMessageQueue)
	if c.reliableOrder != 1 {
		t.Errorf("reliableOrder = %d, want 1 after flushing 0", c.reliableOrder)
	}

	// Send 1, then 0 again (duplicate)
	c.flushPending(createTestPacket(1, "B"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(0, "A-dup"), &s.ClientMessageQueue)

	if c.reliableOrder != 2 {
		t.Errorf("reliableOrder = %d, want 2 after flushing 1", c.reliableOrder)
	}
}

func TestFlushPending_MultipleSequentialBatches(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// First batch: 0, 1, 2
	for i := uint64(0); i <= 2; i++ {
		c.flushPending(createTestPacket(i, string(rune('A'+i))), &s.ClientMessageQueue)
	}
	if c.reliableOrder != 3 {
		t.Errorf("after first batch: reliableOrder = %d, want 3", c.reliableOrder)
	}

	// Second batch with gap: 5, 3, 4
	c.flushPending(createTestPacket(5, "F"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(3, "D"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(4, "E"), &s.ClientMessageQueue)
	if c.reliableOrder != 6 {
		t.Errorf("after second batch: reliableOrder = %d, want 6", c.reliableOrder)
	}
	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer, got %d", len(c.reliableBuffer))
	}
}

func TestFlushPending_SkipAlreadyProcessed(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 5}

	// Send packets 0 through 4 (all below reliableOrder)
	for i := uint64(0); i < 5; i++ {
		c.flushPending(createTestPacket(i, string(rune('A'+i))), &s.ClientMessageQueue)
	}

	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected 0 buffered (all below reliableOrder), got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 5 {
		t.Errorf("reliableOrder should not have changed: got %d, want 5", c.reliableOrder)
	}
}
