/******************************************************************************/
/* network_packet_udp_test.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package network

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestPacketRoundTrip(t *testing.T) {
	original := NetworkPacketUDP{
		timestamp:  1234567890,
		order:      42,
		messageLen: 5,
		typeFlags:  udpPacketTypeReliable,
	}
	copy(original.message[:], []byte("hello"))

	buffer := make([]byte, maxPacketSize)
	n, err := packetToMessage(original, buffer)
	if err != nil {
		t.Fatalf("packetToMessage unexpected error: %v", err)
	}
	deserialized := packetFromMessage(buffer[:n])

	if deserialized.timestamp != original.timestamp {
		t.Errorf("timestamp mismatch: got %d, want %d", deserialized.timestamp, original.timestamp)
	}
	if deserialized.order != original.order {
		t.Errorf("order mismatch: got %d, want %d", deserialized.order, original.order)
	}
	if deserialized.messageLen != original.messageLen {
		t.Errorf("messageLen mismatch: got %d, want %d", deserialized.messageLen, original.messageLen)
	}
	if deserialized.typeFlags != original.typeFlags {
		t.Errorf("typeFlags mismatch: got %d, want %d", deserialized.typeFlags, original.typeFlags)
	}
	if string(deserialized.message[:deserialized.messageLen]) != string(original.message[:original.messageLen]) {
		t.Errorf("message content mismatch: got %q, want %q",
			string(deserialized.message[:deserialized.messageLen]),
			string(original.message[:original.messageLen]))
	}
}

func TestPacketToMessage_BufferTooSmall(t *testing.T) {
	packet := NetworkPacketUDP{
		timestamp:  100,
		order:      1,
		messageLen: 50,
		typeFlags:  udpPacketTypeReliable,
	}
	smallBuffer := make([]byte, 10)
	_, err := packetToMessage(packet, smallBuffer)
	if err == nil {
		t.Error("expected error for buffer too small, got nil")
	}
}

func TestPacketFlags(t *testing.T) {
	tests := []struct {
		name           string
		flags          udpPacketTypeFlags
		expectReliable bool
		expectAck      bool
	}{
		{"none", 0, false, false},
		{"reliable-only", udpPacketTypeReliable, true, false},
		{"ack-only", udpPacketTypeAck, false, true},
		{"both", udpPacketTypeReliable | udpPacketTypeAck, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NetworkPacketUDP{typeFlags: tt.flags}
			if p.isReliable() != tt.expectReliable {
				t.Errorf("isReliable() = %v, want %v (flags=%d)", p.isReliable(), tt.expectReliable, tt.flags)
			}
			if p.isAck() != tt.expectAck {
				t.Errorf("isAck() = %v, want %v (flags=%d)", p.isAck(), tt.expectAck, tt.flags)
			}
		})
	}
}

func TestPacketClone(t *testing.T) {
	original := NetworkPacketUDP{
		timestamp:  999,
		order:      7,
		messageLen: 4,
		typeFlags:  udpPacketTypeAck,
	}
	copy(original.message[:], []byte("test"))

	cloned := original.clone()

	if cloned != original {
		t.Error("clone did not produce equal packet")
	}

	copy(cloned.message[:], []byte("XXXX"))
	cloned.order = 999
	cloned.typeFlags = udpPacketTypeReliable

	if string(original.message[:original.messageLen]) != "test" {
		t.Error("original message was mutated by clone modification")
	}
	if original.order != 7 {
		t.Error("original order was mutated by clone modification")
	}
	if original.typeFlags != udpPacketTypeAck {
		t.Error("original typeFlags were mutated by clone modification")
	}
}

func TestEmptyMessageRoundTrip(t *testing.T) {
	packet := NetworkPacketUDP{
		timestamp:  0,
		order:      0,
		messageLen: 0,
		typeFlags:  0,
	}
	buffer := make([]byte, maxPacketSize)
	n, err := packetToMessage(packet, buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := packetFromMessage(buffer[:n])
	if got.timestamp != 0 || got.order != 0 || got.messageLen != 0 || got.typeFlags != 0 {
		t.Errorf("empty message round-trip failed: %+v", got)
	}
}

func TestMaxPacketMessageSize(t *testing.T) {
	msg := make([]byte, maxPacketSize)
	for i := range msg {
		msg[i] = byte(i % 256)
	}
	packet := NetworkPacketUDP{
		timestamp:  9999,
		order:      1,
		messageLen: maxPacketSize,
		typeFlags:  udpPacketTypeReliable,
	}
	copy(packet.message[:], msg)

	// Need a buffer larger than maxPacketSize because serialized packet includes headers
	bufSize := int(unsafe.Sizeof(packet.timestamp) +
		unsafe.Sizeof(packet.order) +
		unsafe.Sizeof(packet.messageLen) +
		uintptr(packet.messageLen) +
		unsafe.Sizeof(packet.typeFlags))
	buffer := make([]byte, bufSize)
	n, err := packetToMessage(packet, buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := packetFromMessage(buffer[:n])

	if got.messageLen != maxPacketSize {
		t.Errorf("messageLen = %d, want %d", got.messageLen, maxPacketSize)
	}
	for i := uint16(0); i < maxPacketSize; i++ {
		if got.message[i] != msg[i] {
			t.Errorf("byte[%d] = %d, want %d", i, got.message[i], msg[i])
			break
		}
	}
}

func TestPacketToMessageExactBufferSize(t *testing.T) {
	packet := NetworkPacketUDP{
		timestamp:  1,
		order:      2,
		messageLen: 3,
		typeFlags:  udpPacketTypeAck,
	}
	copy(packet.message[:], []byte("abc"))

	totalSize := int(unsafe.Sizeof(packet.timestamp) +
		unsafe.Sizeof(packet.order) +
		unsafe.Sizeof(packet.messageLen) +
		uintptr(packet.messageLen) +
		unsafe.Sizeof(packet.typeFlags))

	exactBuffer := make([]byte, totalSize)
	n, err := packetToMessage(packet, exactBuffer)
	if err != nil {
		t.Fatalf("unexpected error with exact buffer: %v", err)
	}
	if n != totalSize {
		t.Errorf("written size = %d, want %d", n, totalSize)
	}

	got := packetFromMessage(exactBuffer[:n])
	if got.messageLen != packet.messageLen {
		t.Errorf("messageLen mismatch: got %d, want %d", got.messageLen, packet.messageLen)
	}
	if string(got.message[:got.messageLen]) != "abc" {
		t.Errorf("message mismatch: got %q, want %q", string(got.message[:got.messageLen]), "abc")
	}
}

func TestPacketMessageBinaryFormat(t *testing.T) {
	packet := NetworkPacketUDP{
		timestamp:  0x0102030405060708,
		order:      0x090A0B0C0D0E0F10,
		messageLen: 4,
		typeFlags:  udpPacketTypeReliable | udpPacketTypeAck,
	}
	copy(packet.message[:], []byte{0xAA, 0xBB, 0xCC, 0xDD})

	buffer := make([]byte, maxPacketSize)
	_, _ = packetToMessage(packet, buffer)

	expected := []byte{
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
		0x10, 0x0F, 0x0E, 0x0D, 0x0C, 0x0B, 0x0A, 0x09,
		0x04, 0x00,
		0xAA, 0xBB, 0xCC, 0xDD,
		0x03, 0x00, 0x00, 0x00,
	}
	for i, b := range expected {
		if buffer[i] != b {
			t.Errorf("byte[%d] = 0x%02x, want 0x%02x", i, buffer[i], b)
		}
	}
}

func TestPacketFromMessage_PartialRead(t *testing.T) {
	original := NetworkPacketUDP{
		timestamp:  42,
		order:      7,
		messageLen: 2,
		typeFlags:  udpPacketTypeAck,
	}
	copy(original.message[:], []byte{0xDE, 0xAD})

	buffer := make([]byte, maxPacketSize)
	n, _ := packetToMessage(original, buffer)

	extraBuffer := make([]byte, n+100)
	copy(extraBuffer, buffer[:n])
	copy(extraBuffer[n:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	got := packetFromMessage(extraBuffer[:n])
	if got.messageLen != 2 {
		t.Errorf("messageLen = %d, want 2", got.messageLen)
	}
	if got.message[0] != 0xDE || got.message[1] != 0xAD {
		t.Error("message content corrupted by trailing garbage")
	}
}

func TestPacketRoundTrip_AllFlagCombinations(t *testing.T) {
	flags := []udpPacketTypeFlags{
		0,
		udpPacketTypeReliable,
		udpPacketTypeAck,
		udpPacketTypeReliable | udpPacketTypeAck,
		0xFF,
	}
	for _, f := range flags {
		packet := NetworkPacketUDP{
			timestamp:  1,
			order:      2,
			messageLen: 1,
			typeFlags:  f,
		}
		packet.message[0] = 0x42

		buf := make([]byte, maxPacketSize)
		n, err := packetToMessage(packet, buf)
		if err != nil {
			t.Errorf("flags=0x%x: unexpected error: %v", f, err)
			continue
		}
		got := packetFromMessage(buf[:n])
		if got.typeFlags != f {
			t.Errorf("flags=0x%x: typeFlags round-trip failed, got 0x%x", f, got.typeFlags)
		}
	}
}

func TestPacketTimestampExtraction(t *testing.T) {
	packet := NetworkPacketUDP{
		timestamp:  123456789,
		messageLen: 8,
	}
	binary.LittleEndian.PutUint64(packet.message[:], 123456789)

	buf := make([]byte, maxPacketSize)
	n, _ := packetToMessage(packet, buf)
	_ = n

	tsBytes := buf[unsafe.Sizeof(packet.timestamp)+unsafe.Sizeof(packet.order)+unsafe.Sizeof(packet.messageLen):]
	extracted := binary.LittleEndian.Uint64(tsBytes)
	if extracted != 123456789 {
		t.Errorf("extracted timestamp = %d, want 123456789", extracted)
	}
}
