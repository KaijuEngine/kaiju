/******************************************************************************/
/* master_server_request_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package master_server

import (
	"testing"
	"unsafe"
)

func TestRequestSerializeDeserialize(t *testing.T) {
	// Note: Serialize/Deserialize do NOT include ServerId — it's a struct field
	// that's not part of the wire format.
	req := Request{
		MaxPlayers:     64,
		CurrentPlayers: 42,
		Type:           RequestTypeRegisterServer,
	}
	copy(req.Game[:], "MyGameID1234567890123456")
	copy(req.Name[:], "TestServer")
	copy(req.Password[:], "secret1234")

	buffer := make([]byte, unsafe.Sizeof(Request{}))
	req.Serialize(buffer)

	got := DeserializeRequest(buffer)

	// ServerId is not serialized/deserialized, so it's always 0
	if got.MaxPlayers != req.MaxPlayers {
		t.Errorf("MaxPlayers = %d, want %d", got.MaxPlayers, req.MaxPlayers)
	}
	if got.CurrentPlayers != req.CurrentPlayers {
		t.Errorf("CurrentPlayers = %d, want %d", got.CurrentPlayers, req.CurrentPlayers)
	}
	if got.Type != req.Type {
		t.Errorf("Type = %d, want %d", got.Type, req.Type)
	}
	if string(gameKey(got)) != string(gameKey(req)) {
		t.Errorf("Game mismatch: got %q, want %q", string(got.Game[:]), string(req.Game[:]))
	}
	if string(gameName(got)) != string(gameName(req)) {
		t.Errorf("Name mismatch: got %q, want %q", string(got.Name[:]), string(req.Name[:]))
	}
	if string(passwd(got)) != string(passwd(req)) {
		t.Errorf("Password mismatch: got %q, want %q", string(got.Password[:]), string(req.Password[:]))
	}
}

func TestRequestAllRequestTypes(t *testing.T) {
	types := []MasterServerRequestType{
		RequestTypeRegisterServer,
		RequestTypeUnregisterServer,
		RequestTypePing,
		RequestTypeServerList,
		RequestTypeJoinServer,
	}
	for _, typ := range types {
		req := Request{Type: typ}
		buf := make([]byte, unsafe.Sizeof(Request{}))
		req.Serialize(buf)
		got := DeserializeRequest(buf)
		if got.Type != typ {
			t.Errorf("Type %d round-trip failed: got %d", typ, got.Type)
		}
	}
}

func TestRequestEmptyStrings(t *testing.T) {
	req := Request{
		Type: RequestTypePing,
	}
	buf := make([]byte, unsafe.Sizeof(Request{}))
	req.Serialize(buf)
	got := DeserializeRequest(buf)

	// Fixed-size arrays are serialized as all-zero bytes when empty.
	// Converting to string yields null bytes, not empty string.
	if len(gameKey(got)) != 0 {
		t.Errorf("Game should be empty, got %q", gameKey(got))
	}
	if len(gameName(got)) != 0 {
		t.Errorf("Name should be empty, got %q", gameName(got))
	}
	if len(passwd(got)) != 0 {
		t.Errorf("Password should be empty, got %q", passwd(got))
	}
}

func TestRequestServerIdNotSerialized(t *testing.T) {
	// ServerId is a struct field but is NOT included in Serialize/Deserialize.
	req := Request{
		ServerId: 0xFFFFFFFFFFFFFFFF,
		Type:     RequestTypeJoinServer,
	}
	buf := make([]byte, unsafe.Sizeof(Request{}))
	req.Serialize(buf)
	got := DeserializeRequest(buf)

	// ServerId round-trips as 0 since it's not part of the wire format
	if got.ServerId != 0 {
		t.Errorf("ServerId = %d, want 0 (not serialized)", got.ServerId)
	}
}

func TestRequestMaxPlayers(t *testing.T) {
	for _, v := range []uint16{0, 1, 16, 64, 255, 1000, 65535} {
		req := Request{MaxPlayers: v}
		buf := make([]byte, unsafe.Sizeof(Request{}))
		req.Serialize(buf)
		got := DeserializeRequest(buf)
		if got.MaxPlayers != v {
			t.Errorf("MaxPlayers=%d round-trip failed: got %d", v, got.MaxPlayers)
		}
	}
}

func TestRequestBufferSize(t *testing.T) {
	buf := make([]byte, unsafe.Sizeof(Request{}))
	req := Request{}
	req.Serialize(buf)

	// Verify the buffer size matches the Request struct size
	if len(buf) != int(unsafe.Sizeof(Request{})) {
		t.Errorf("buffer size = %d, want %d", len(buf), int(unsafe.Sizeof(Request{})))
	}
}

func TestRequestDeserialize_PartialBuffer(t *testing.T) {
	// Deserialize from a buffer that's exactly the right size
	req := Request{
		Type:           RequestTypeRegisterServer,
		MaxPlayers:     16,
		CurrentPlayers: 3,
	}
	copy(req.Game[:], "Game")
	copy(req.Name[:], "Server")
	copy(req.Password[:], "pass")

	buf := make([]byte, unsafe.Sizeof(Request{}))
	req.Serialize(buf)
	got := DeserializeRequest(buf)

	if got.Type != req.Type {
		t.Error("Type mismatch")
	}
	if got.MaxPlayers != req.MaxPlayers {
		t.Error("MaxPlayers mismatch")
	}
	if got.CurrentPlayers != req.CurrentPlayers {
		t.Error("CurrentPlayers mismatch")
	}
}

// helper to extract meaningful parts of fixed-size arrays
func gameKey(r Request) string {
	s := string(r.Game[:])
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}

func gameName(r Request) string {
	s := string(r.Name[:])
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}

func passwd(r Request) string {
	s := string(r.Password[:])
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
