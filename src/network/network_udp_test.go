/******************************************************************************/
/* network_udp_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package network

import (
	"testing"
	"time"
)

func TestCreateUnreliablePacket(t *testing.T) {
	n := NetworkUDP{}
	msg := []byte("hello world")

	packet := n.createUnreliable(msg)

	if packet.messageLen != uint16(len(msg)) {
		t.Errorf("messageLen = %d, want %d", packet.messageLen, len(msg))
	}
	if string(packet.message[:packet.messageLen]) != string(msg) {
		t.Errorf("message content mismatch: got %q, want %q",
			string(packet.message[:packet.messageLen]), string(msg))
	}
	if packet.isReliable() {
		t.Error("unreliable packet should not have reliable flag")
	}
	if packet.isAck() {
		t.Error("unreliable packet should not have ack flag")
	}
	if packet.order != 0 {
		t.Errorf("order = %d, want 0", packet.order)
	}
	if packet.timestamp == 0 {
		t.Error("timestamp should be non-zero")
	}
	// Unreliable packets have zero nextRetry (no retry logic)
	if !packet.nextRetry.IsZero() {
		t.Error("nextRetry should be zero for unreliable packets (no retry)")
	}
}

func TestCreateReliablePacket_IncrementsOrder(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}
	msg := []byte("reliable msg")

	_ = n.createReliable(msg, client)
	_ = n.createReliable(msg, client)
	_ = n.createReliable(msg, client)

	if client.reliableOrder != 3 {
		t.Errorf("reliableOrder = %d, want 3", client.reliableOrder)
	}
	if len(n.pendingPackets) != 3 {
		t.Errorf("pendingPackets count = %d, want 3", len(n.pendingPackets))
	}
}

func TestCreateReliablePacket_HasReliableFlag(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}
	msg := []byte("test")

	packet := n.createReliable(msg, client)

	if !packet.isReliable() {
		t.Error("reliable packet should have reliable flag")
	}
	if packet.isAck() {
		t.Error("reliable packet should not have ack flag")
	}
	if packet.order != 0 {
		t.Errorf("order = %d, want 0", packet.order)
	}
	if packet.timestamp == 0 {
		t.Error("timestamp should be non-zero")
	}
	if packet.nextRetry.IsZero() {
		t.Error("nextRetry should be set for reliable packets")
	}
	if !packet.nextRetry.After(time.Now().Add(-time.Second)) {
		t.Error("nextRetry should be in the future")
	}
}

func TestCreateReliablePacket_StartingOrder(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 10}
	msg := []byte("test")

	packet := n.createReliable(msg, client)

	if packet.order != 10 {
		t.Errorf("packet order = %d, want 10", packet.order)
	}
	if client.reliableOrder != 11 {
		t.Errorf("client reliableOrder = %d, want 11", client.reliableOrder)
	}
}

func TestCreateAckPacket(t *testing.T) {
	n := NetworkUDP{}
	timestampBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	packet := n.createAck(timestampBytes)

	if !packet.isAck() {
		t.Error("ack packet should have ack flag")
	}
	if packet.isReliable() {
		t.Error("ack packet should not have reliable flag")
	}
	if packet.messageLen != uint16(len(timestampBytes)) {
		t.Errorf("messageLen = %d, want %d", packet.messageLen, len(timestampBytes))
	}
	for i, b := range timestampBytes {
		if packet.message[i] != b {
			t.Errorf("message[%d] = 0x%02x, want 0x%02x", i, packet.message[i], b)
			break
		}
	}
}

func TestRemovePendingPacket(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}

	_ = n.createReliable([]byte("a"), client)
	time.Sleep(time.Millisecond)
	_ = n.createReliable([]byte("b"), client)
	time.Sleep(time.Millisecond)
	_ = n.createReliable([]byte("c"), client)

	if len(n.pendingPackets) != 3 {
		t.Fatalf("expected 3 pending, got %d", len(n.pendingPackets))
	}

	// Remove the middle one
	targetTimestamp := n.pendingPackets[1].packet.timestamp
	n.removePendingPacket(targetTimestamp)

	if len(n.pendingPackets) != 2 {
		t.Errorf("expected 2 pending after removal, got %d", len(n.pendingPackets))
	}

	// Verify the correct one was removed
	for _, pp := range n.pendingPackets {
		if pp.packet.timestamp == targetTimestamp {
			t.Error("removed packet still exists in pending")
		}
	}
}

func TestRemovePendingPacket_EmptyList(t *testing.T) {
	n := NetworkUDP{}
	n.removePendingPacket(12345)
	if len(n.pendingPackets) != 0 {
		t.Error("should not panic on empty list")
	}
}

func TestRemovePendingPacket_NonExistentTimestamp(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}
	_ = n.createReliable([]byte("a"), client)

	n.removePendingPacket(999999)
	if len(n.pendingPackets) != 1 {
		t.Errorf("expected 1 pending after removing non-existent, got %d", len(n.pendingPackets))
	}
}

func TestRemovePendingPacket_FirstElement(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}

	p1 := n.createReliable([]byte("a"), client)
	p2 := n.createReliable([]byte("b"), client)

	n.removePendingPacket(p1.timestamp)

	if len(n.pendingPackets) != 1 {
		t.Errorf("expected 1 pending, got %d", len(n.pendingPackets))
	}
	if n.pendingPackets[0].packet.timestamp != p2.timestamp {
		t.Error("wrong packet remained after removing first")
	}
}

func TestRemovePendingPacket_LastElement(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}

	p1 := n.createReliable([]byte("a"), client)
	p2 := n.createReliable([]byte("b"), client)

	n.removePendingPacket(p2.timestamp)

	if len(n.pendingPackets) != 1 {
		t.Errorf("expected 1 pending, got %d", len(n.pendingPackets))
	}
	if n.pendingPackets[0].packet.timestamp != p1.timestamp {
		t.Error("wrong packet remained after removing last")
	}
}

func TestIsLive(t *testing.T) {
	n := NetworkUDP{}
	if n.IsLive() {
		t.Error("expected IsLive() = false for closed connection")
	}
}

func TestReliableRetryDelay(t *testing.T) {
	if reliableRetryDelay <= 0 {
		t.Error("reliableRetryDelay should be positive")
	}
	if reliableRetryDelay != time.Millisecond*15 {
		t.Errorf("reliableRetryDelay = %v, want %v", reliableRetryDelay, time.Millisecond*15)
	}
}

func TestPendingPacketTargetReference(t *testing.T) {
	n := NetworkUDP{}
	client := &ServerClient{reliableOrder: 0}

	_ = n.createReliable([]byte("test"), client)

	if len(n.pendingPackets) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(n.pendingPackets))
	}
	if n.pendingPackets[0].target != client {
		t.Error("pending packet target should reference the same client")
	}
}

func TestCreateUnreliable_EmptyMessage(t *testing.T) {
	n := NetworkUDP{}
	packet := n.createUnreliable([]byte{})

	if packet.messageLen != 0 {
		t.Errorf("messageLen = %d, want 0", packet.messageLen)
	}
	if packet.timestamp == 0 {
		t.Error("timestamp should still be set even for empty message")
	}
}

func TestCreateReliable_MultipleClients(t *testing.T) {
	n := NetworkUDP{}
	client1 := &ServerClient{reliableOrder: 0}
	client2 := &ServerClient{reliableOrder: 5}

	p1 := n.createReliable([]byte("for client1"), client1)
	p2 := n.createReliable([]byte("for client2"), client2)

	if p1.order != 0 {
		t.Errorf("p1 order = %d, want 0", p1.order)
	}
	if p2.order != 5 {
		t.Errorf("p2 order = %d, want 5", p2.order)
	}
	if client1.reliableOrder != 1 {
		t.Errorf("client1 reliableOrder = %d, want 1", client1.reliableOrder)
	}
	if client2.reliableOrder != 6 {
		t.Errorf("client2 reliableOrder = %d, want 6", client2.reliableOrder)
	}
	if len(n.pendingPackets) != 2 {
		t.Errorf("pendingPackets = %d, want 2", len(n.pendingPackets))
	}
	if n.pendingPackets[0].target != client1 || n.pendingPackets[1].target != client2 {
		t.Error("pending packet targets mismatch")
	}
}
