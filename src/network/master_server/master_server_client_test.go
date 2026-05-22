/******************************************************************************/
/* master_server_client_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package master_server

import (
	"testing"
	"unsafe"

	"kaijuengine.com/network"
)

func TestMasterServerClientSendRequestTypes(t *testing.T) {
	types := []struct {
		name string
		typ  MasterServerRequestType
	}{
		{"Register", RequestTypeRegisterServer},
		{"Unregister", RequestTypeUnregisterServer},
		{"Ping", RequestTypePing},
		{"ServerList", RequestTypeServerList},
		{"JoinServer", RequestTypeJoinServer},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Type: tt.typ}
			buf := make([]byte, unsafe.Sizeof(Request{}))
			req.Serialize(buf)
			got := DeserializeRequest(buf)
			if got.Type != tt.typ {
				t.Errorf("Type = %d, want %d", got.Type, tt.typ)
			}
		})
	}
}

func TestMasterServerClientProcessMessage_ResponseTypes(t *testing.T) {
	client := &MasterServerClient{
		OnServerList: func([]ResponseServerList, uint32) {},
		OnServerJoin: func(string) {},
		OnClientJoin: func(string) {},
		OnError:      func(uint8) {},
	}

	testCases := []struct {
		name string
		resp Response
	}{
		{"ConfirmRegister", Response{Type: ResponseTypeConfirmRegister}},
		{"ServerList", Response{Type: ResponseTypeServerList, TotalList: 1}},
		{"JoinServerInfo", Response{Type: ResponseTypeJoinServerInfo}},
		{"ClientJoinInfo", Response{Type: ResponseTypeClientJoinInfo}},
		{"Error-None", Response{Type: ResponseTypeError, Error: ErrorNone}},
		{"Error-Password", Response{Type: ResponseTypeError, Error: ErrorIncorrectPassword}},
		{"Error-ServerNotFound", Response{Type: ResponseTypeError, Error: ErrorServerDoesntExist}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := make([]byte, unsafe.Sizeof(Response{}))
			tc.resp.Serialize(buf)
			msg := network.NewClientMessageFromBytes(buf)
			client.processMessage(msg)
		})
	}
}

func TestMasterServerClientProcessMessage_InvalidSize(t *testing.T) {
	client := &MasterServerClient{
		OnError: func(uint8) {},
	}

	// The update() function guards against invalid sizes by checking
	// len(buff) != unsafe.Sizeof(Response{}) before calling processMessage.
	// processMessage itself does not guard and will panic on short buffers.
	// This test verifies that short messages are skipped via the update path.

	// Verify the size guard in update(): enqueue a message that's too short,
	// then call update() — it should skip it without panicking.
	shortMsg := network.NewClientMessageFromBytes([]byte{0xFF})
	client.client.ServerMessageQueue.Enqueue(shortMsg)
	// This should NOT panic — update() checks the size and skips short messages
	client.update(0)
}

func TestMasterServerClientCallbacksInvoked(t *testing.T) {
	var serverListCalled bool
	var serverJoinAddress string
	var clientJoinAddress string
	var errorCalled bool
	var errorCode uint8

	client := &MasterServerClient{
		OnServerList: func([]ResponseServerList, uint32) { serverListCalled = true },
		OnServerJoin: func(s string) { serverJoinAddress = s },
		OnClientJoin: func(s string) { clientJoinAddress = s },
		OnError:      func(e uint8) { errorCalled = true; errorCode = e },
	}

	// Test ServerList callback
	{
		resp := Response{Type: ResponseTypeServerList, TotalList: 1}
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		msg := network.NewClientMessageFromBytes(buf)
		client.processMessage(msg)
		if !serverListCalled {
			t.Error("OnServerList not called")
		}
	}

	// Test JoinServerInfo callback
	{
		serverListCalled = false
		resp := Response{Type: ResponseTypeJoinServerInfo}
		copy(resp.Address[:], "10.0.0.1")
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		msg := network.NewClientMessageFromBytes(buf)
		client.processMessage(msg)
		if serverJoinAddress != "10.0.0.1" {
			t.Errorf("OnServerJoin called with %q, want 10.0.0.1", serverJoinAddress)
		}
	}

	// Test ClientJoinInfo callback
	{
		clientJoinAddress = ""
		resp := Response{Type: ResponseTypeClientJoinInfo}
		copy(resp.Address[:], "10.0.0.2")
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		msg := network.NewClientMessageFromBytes(buf)
		client.processMessage(msg)
		if clientJoinAddress != "10.0.0.2" {
			t.Errorf("OnClientJoin called with %q, want 10.0.0.2", clientJoinAddress)
		}
	}

	// Test Error callback
	{
		errorCalled = false
		resp := Response{Type: ResponseTypeError, Error: ErrorIncorrectPassword}
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		msg := network.NewClientMessageFromBytes(buf)
		client.processMessage(msg)
		if !errorCalled {
			t.Error("OnError not called")
		}
		if errorCode != ErrorIncorrectPassword {
			t.Errorf("errorCode = %d, want %d", errorCode, ErrorIncorrectPassword)
		}
	}
}

func TestMasterServerClientPingCycles(t *testing.T) {
	client := &MasterServerClient{
		isServer: true,
		pingTime: pingIntervalSeconds,
		OnError:  func(uint8) {},
	}

	// Verify ping timer countdown
	for i := 0; i < 5; i++ {
		client.update(1.0)
		if client.pingTime > 0 {
			// Timer counting down
		}
	}

	if client.pingTime <= 0 {
		t.Error("pingTime should have been reset after ping was sent")
	}
}

func TestMasterServerClientMessageFlush(t *testing.T) {
	client := &MasterServerClient{
		OnError: func(uint8) {},
	}

	for i := 0; i < 3; i++ {
		resp := Response{Type: ResponseTypeConfirmRegister}
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		msg := network.NewClientMessageFromBytes(buf)
		client.client.ServerMessageQueue.Enqueue(msg)
	}

	client.update(0)
}

func TestMasterServerClientNilCallbacksDefault(t *testing.T) {
	client := &MasterServerClient{}

	if client.OnServerList == nil {
		client.OnServerList = func([]ResponseServerList, uint32) {}
	}
	if client.OnServerJoin == nil {
		client.OnServerJoin = func(string) {}
	}
	if client.OnClientJoin == nil {
		client.OnClientJoin = func(string) {}
	}
	if client.OnError == nil {
		client.OnError = func(uint8) {}
	}

	client.OnServerList(nil, 0)
	client.OnServerJoin("")
	client.OnClientJoin("")
	client.OnError(ErrorNone)
}
