/******************************************************************************/
/* master_server_logic_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package master_server

import (
	"testing"
	"unsafe"
)

func TestDeserializeRequest_ExactSize(t *testing.T) {
	reqSize := int(unsafe.Sizeof(Request{}))

	// DeserializeRequest does not guard against short buffers and will panic.
	// Only test the exact-size path.
	exact := make([]byte, reqSize)
	req := Request{Type: RequestTypePing}
	req.Serialize(exact)
	got := DeserializeRequest(exact)
	if got.Type != RequestTypePing {
		t.Errorf("Type = %d, want %d", got.Type, RequestTypePing)
	}
}

func TestRequestResponseSizeComparison(t *testing.T) {
	reqSize := int(unsafe.Sizeof(Request{}))
	respSize := int(unsafe.Sizeof(Response{}))

	if reqSize <= 0 {
		t.Error("Request size should be positive")
	}
	if respSize <= 0 {
		t.Error("Response size should be positive")
	}

	t.Logf("Request struct size: %d bytes", reqSize)
	t.Logf("Response struct size: %d bytes", respSize)

	// Response should be larger due to server list array
	if respSize <= reqSize {
		t.Errorf("Response size (%d) should be larger than Request size (%d)", respSize, reqSize)
	}
}

func TestRequestSerializationOrder(t *testing.T) {
	// Serialize order: Game(32), Name(64), Password(16), MaxPlayers(2), CurrentPlayers(2), Type(1)
	// ServerId is NOT serialized despite being a struct field.
	req := Request{
		MaxPlayers:     100,
		CurrentPlayers: 50,
		Type:           RequestTypeRegisterServer,
	}
	copy(req.Game[:], "GameKey")
	copy(req.Name[:], "ServerName")
	copy(req.Password[:], "Password123")

	buf := make([]byte, 117) // Game(32) + Name(64) + Password(16) + MaxPlayers(2) + CurrentPlayers(2) + Type(1)
	req.Serialize(buf)

	// Game at offset 0 (32 bytes) — use bytes.Equal to avoid null byte issues
	if string(buf[0:7]) != "GameKey" {
		t.Errorf("Game not at expected offset: got %q", string(buf[0:7]))
	}

	// Name at offset 32 (64 bytes)
	if string(buf[32:42]) != "ServerName" {
		t.Errorf("Name not at expected offset: got %q", string(buf[32:42]))
	}

	// Password at offset 96 (16 bytes)
	if string(buf[96:107]) != "Password123" {
		t.Errorf("Password not at expected offset: got %q", string(buf[96:107]))
	}

	// MaxPlayers at offset 112 (2 bytes)
	if buf[112] != 100 || buf[113] != 0 {
		t.Errorf("MaxPlayers not at expected offset: buf[112]=%d, buf[113]=%d", buf[112], buf[113])
	}

	// CurrentPlayers at offset 114 (2 bytes)
	if buf[114] != 50 || buf[115] != 0 {
		t.Errorf("CurrentPlayers not at expected offset: buf[114]=%d, buf[115]=%d", buf[114], buf[115])
	}

	// Type at offset 116 (1 byte)
	if buf[116] != byte(RequestTypeRegisterServer) {
		t.Errorf("Type not at expected offset: got %d", buf[116])
	}
}

func TestResponseSerializationOrder(t *testing.T) {
	resp := Response{
		Type:      ResponseTypeServerList,
		TotalList: 5,
		Error:     ErrorNone,
	}
	copy(resp.Address[:], "192.168.1.1")

	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buf)

	// Type should be first byte
	if buf[0] != uint8(ResponseTypeServerList) {
		t.Errorf("Type byte = %d, want %d", buf[0], ResponseTypeServerList)
	}
}

func TestMasterServerClientPingInterval(t *testing.T) {
	if pingIntervalSeconds != 3.0 {
		t.Errorf("pingIntervalSeconds = %f, want 3.0", pingIntervalSeconds)
	}
}

func TestMasterServerPort(t *testing.T) {
	if masterPort != 15973 {
		t.Errorf("masterPort = %d, want 15973", masterPort)
	}
}

func TestServerTimeout(t *testing.T) {
	if serverTimeout.Seconds() != 30 {
		t.Errorf("serverTimeout = %v, want 30s", serverTimeout)
	}
}

func TestMasterServerAddress(t *testing.T) {
	if masterAddress != "localhost" {
		t.Errorf("masterAddress = %q, want %q", masterAddress, "localhost")
	}
}

func TestMasterServerClientCallbacksNil(t *testing.T) {
	client := &MasterServerClient{}
	if client.OnServerList != nil {
		t.Error("OnServerList should be nil by default")
	}
	if client.OnServerJoin != nil {
		t.Error("OnServerJoin should be nil by default")
	}
	if client.OnClientJoin != nil {
		t.Error("OnClientJoin should be nil by default")
	}
	if client.OnError != nil {
		t.Error("OnError should be nil by default")
	}
}

func TestMasterServerClientPingTimeDecrement(t *testing.T) {
	client := &MasterServerClient{
		isServer: true,
		pingTime: pingIntervalSeconds,
	}
	client.update(1.0)
	if client.pingTime != 2.0 {
		t.Errorf("pingTime = %f, want 2.0", client.pingTime)
	}
	client.update(2.0)
	if client.pingTime != 0.0 {
		t.Errorf("pingTime = %f, want 0.0", client.pingTime)
	}
}

func TestMasterServerClientNonServerNoPing(t *testing.T) {
	client := &MasterServerClient{
		isServer: false,
		pingTime: pingIntervalSeconds,
	}
	client.update(pingIntervalSeconds + 1)
	if client.pingTime != pingIntervalSeconds {
		t.Errorf("pingTime should not change for non-server: got %f", client.pingTime)
	}
}
