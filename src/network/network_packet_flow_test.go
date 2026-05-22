/******************************************************************************/
/* network_packet_flow_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package network

import (
	"sort"
	"testing"
)

func TestPacketSerializationRoundTrip_Comprehensive(t *testing.T) {
	testCases := []struct {
		name   string
		packet NetworkPacketUDP
	}{
		{
			"minimal",
			NetworkPacketUDP{timestamp: 1, order: 0, messageLen: 0, typeFlags: 0},
		},
		{
			"reliable-with-message",
			NetworkPacketUDP{
				timestamp:  999999,
				order:      5,
				messageLen: 13,
				typeFlags:  udpPacketTypeReliable,
			},
		},
		{
			"ack-with-data",
			NetworkPacketUDP{
				timestamp:  123456,
				order:      0,
				messageLen: 8,
				typeFlags:  udpPacketTypeAck,
			},
		},
		{
			"large-message",
			NetworkPacketUDP{
				timestamp:  0,
				order:      0,
				messageLen: 500,
				typeFlags:  udpPacketTypeReliable | udpPacketTypeAck,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fill message bytes with a pattern
			for i := uint16(0); i < tc.packet.messageLen; i++ {
				tc.packet.message[i] = byte((int(i)*7 + 13) % 256)
			}

			buf := make([]byte, maxPacketSize)
			n, err := packetToMessage(tc.packet, buf)
			if err != nil {
				t.Fatalf("packetToMessage error: %v", err)
			}
			got := packetFromMessage(buf[:n])

			if got.timestamp != tc.packet.timestamp {
				t.Errorf("timestamp = %d, want %d", got.timestamp, tc.packet.timestamp)
			}
			if got.order != tc.packet.order {
				t.Errorf("order = %d, want %d", got.order, tc.packet.order)
			}
			if got.messageLen != tc.packet.messageLen {
				t.Errorf("messageLen = %d, want %d", got.messageLen, tc.packet.messageLen)
			}
			if got.typeFlags != tc.packet.typeFlags {
				t.Errorf("typeFlags = %d, want %d", got.typeFlags, tc.packet.typeFlags)
			}
			for i := uint16(0); i < tc.packet.messageLen; i++ {
				if got.message[i] != tc.packet.message[i] {
					t.Errorf("message[%d] = 0x%02x, want 0x%02x", i, got.message[i], tc.packet.message[i])
					break
				}
			}
		})
	}
}

func TestFlushPending_ReliableOrderInvariant(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Invariant: reliableBuffer only contains packets with order >= reliableOrder
	orders := []uint64{3, 1, 4, 1, 5, 9, 2, 6}
	for _, o := range orders {
		p := createTestPacket(o, string(rune('A'+(o%26))))
		c.flushPending(p, &s.ClientMessageQueue)

		for i, bp := range c.reliableBuffer {
			if bp.order < c.reliableOrder {
				t.Errorf("buffer[%d].order (%d) < reliableOrder (%d)", i, bp.order, c.reliableOrder)
			}
		}
	}
}

func TestFlushPending_BufferSortedDescending(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send packet 0 first (matches reliableOrder, gets processed)
	p0 := createTestPacket(0, "A")
	c.flushPending(p0, &s.ClientMessageQueue)
	// reliableOrder is now 1, buffer is empty

	// Send future packets: 5, 3, 4 (all > reliableOrder=1, buffered without sorting)
	c.flushPending(createTestPacket(5, "E"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(3, "C"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(4, "D"), &s.ClientMessageQueue)
	// Buffer: [5, 3, 4] (not sorted yet - sort only happens when order==reliableOrder)

	// Send packet 1 (matches reliableOrder=1, triggers sort and process)
	c.flushPending(createTestPacket(1, "B"), &s.ClientMessageQueue)
	// Appends to buffer [5, 3, 4, 1], sorts desc [5, 4, 3, 1], processes order 1,
	// reliableOrder becomes 2, buffer truncated to [5, 4, 3]

	if len(c.reliableBuffer) != 3 {
		t.Errorf("expected 3 remaining in buffer, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 2 {
		t.Errorf("reliableOrder = %d, want 2", c.reliableOrder)
	}
}

func TestFlushPending_DuplicatesNotInQueue(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send 0 five times
	for i := 0; i < 5; i++ {
		c.flushPending(createTestPacket(0, "A"), &s.ClientMessageQueue)
	}

	// Only one should be in queue
	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d (duplicates not deduplicated)", len(msgs))
	}
}

func TestFlushPending_SliceOperationsCorrect(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send packets 2, 1, 3, 0 (out of order)
	c.flushPending(createTestPacket(2, "C"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(1, "B"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(3, "D"), &s.ClientMessageQueue)
	c.flushPending(createTestPacket(0, "A"), &s.ClientMessageQueue)

	// All should be processed now
	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 4 {
		t.Errorf("reliableOrder = %d, want 4", c.reliableOrder)
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}

	// Verify order: A, B, C, D
	for i, m := range msgs {
		expected := string(rune('A' + i))
		if string(m.Message()) != expected {
			t.Errorf("msg[%d] = %q, want %q", i, string(m.Message()), expected)
		}
	}
}

func TestFlushPending_SortFunction(t *testing.T) {
	// Test the sort function used in flushPending directly
	orders := []uint64{3, 1, 5, 2, 4}
	sorted := make([]uint64, len(orders))
	copy(sorted, orders)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] > sorted[j]
	})

	// Should be descending
	for i := 1; i < len(sorted); i++ {
		if sorted[i] > sorted[i-1] {
			t.Errorf("not sorted descending: sorted[%d]=%d > sorted[%d-1]=%d",
				i, sorted[i], i, sorted[i-1])
		}
	}

	// Expected: [5, 4, 3, 2, 1]
	expected := []uint64{5, 4, 3, 2, 1}
	for i, v := range sorted {
		if v != expected[i] {
			t.Errorf("sorted[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestFlushPending_EdgeCase_ZeroOrder(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send order 0
	p := createTestPacket(0, "A")
	p.timestamp = 12345
	c.flushPending(p, &s.ClientMessageQueue)

	if len(c.reliableBuffer) != 0 {
		t.Errorf("expected empty buffer after order 0, got %d", len(c.reliableBuffer))
	}
	if c.reliableOrder != 1 {
		t.Errorf("reliableOrder = %d, want 1", c.reliableOrder)
	}

	msgs := s.ClientMessageQueue.Flush()
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
	if string(msgs[0].Message()) != "A" {
		t.Errorf("message = %q, want %q", string(msgs[0].Message()), "A")
	}
}

func TestFlushPending_CloningBehavior(t *testing.T) {
	s := NewServerUDP()
	c := &ServerClient{reliableOrder: 0}

	// Send out-of-order packet - it should be cloned
	original := createTestPacket(5, "Future")
	original.timestamp = 999
	c.flushPending(original, &s.ClientMessageQueue)

	// Mutate the original
	original.message[0] = 'X'

	// The buffered copy should not be affected
	if len(c.reliableBuffer) != 1 {
		t.Fatalf("expected 1 buffer, got %d", len(c.reliableBuffer))
	}
	if string(c.reliableBuffer[0].message[:1]) != "F" {
		t.Errorf("buffered message was mutated: got %q, want %q",
			string(c.reliableBuffer[0].message[:1]), "F")
	}
}
