/******************************************************************************/
/* master_server_response_test.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package master_server

import (
	"testing"
	"unsafe"
)

func TestResponseSerializeDeserialize(t *testing.T) {
	resp := Response{
		Type:      ResponseTypeServerList,
		TotalList: 5,
		Error:     ErrorNone,
	}
	copy(resp.Address[:], "192.168.1.100")

	// Add some server entries
	for i := 0; i < 3; i++ {
		resp.List[i].Id = uint64(i + 100)
		resp.List[i].MaxPlayers = 16
		resp.List[i].CurrentPlayers = uint16(i)
		copy(resp.List[i].Name[:], "Server")
	}

	buffer := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buffer)

	got := DeserializeResponse(buffer)

	if got.Type != resp.Type {
		t.Errorf("Type = %d, want %d", got.Type, resp.Type)
	}
	if got.TotalList != resp.TotalList {
		t.Errorf("TotalList = %d, want %d", got.TotalList, resp.TotalList)
	}
	if got.Error != resp.Error {
		t.Errorf("Error = %d, want %d", got.Error, resp.Error)
	}
	if string(addressTrim(got)) != "192.168.1.100" {
		t.Errorf("Address = %q, want %q", string(addressTrim(got)), "192.168.1.100")
	}

	// Check server list entries
	for i := 0; i < 3; i++ {
		if got.List[i].Id != resp.List[i].Id {
			t.Errorf("List[%d].Id = %d, want %d", i, got.List[i].Id, resp.List[i].Id)
		}
		if got.List[i].MaxPlayers != resp.List[i].MaxPlayers {
			t.Errorf("List[%d].MaxPlayers = %d, want %d", i, got.List[i].MaxPlayers, resp.List[i].MaxPlayers)
		}
		if got.List[i].CurrentPlayers != resp.List[i].CurrentPlayers {
			t.Errorf("List[%d].CurrentPlayers = %d, want %d", i, got.List[i].CurrentPlayers, resp.List[i].CurrentPlayers)
		}
	}
}

func TestResponseServerListFullCapacity(t *testing.T) {
	resp := Response{Type: ResponseTypeServerList, TotalList: 10}

	for i := 0; i < serversPerResponse; i++ {
		resp.List[i].Id = uint64(i)
		resp.List[i].MaxPlayers = uint16(i * 2)
		resp.List[i].CurrentPlayers = uint16(i)
	}

	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buf)

	got := DeserializeResponse(buf)

	for i := 0; i < serversPerResponse; i++ {
		if got.List[i].Id != uint64(i) {
			t.Errorf("List[%d].Id = %d, want %d", i, got.List[i].Id, i)
		}
		if got.List[i].MaxPlayers != uint16(i*2) {
			t.Errorf("List[%d].MaxPlayers = %d, want %d", i, got.List[i].MaxPlayers, i*2)
		}
	}
}

func TestResponseAllTypes(t *testing.T) {
	types := []MasterServerResponseType{
		ResponseTypeConfirmRegister,
		ResponseTypeServerList,
		ResponseTypeJoinServerInfo,
		ResponseTypeClientJoinInfo,
		ResponseTypeError,
	}
	for _, typ := range types {
		resp := Response{Type: typ}
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		got := DeserializeResponse(buf)
		if got.Type != typ {
			t.Errorf("ResponseType %d round-trip failed: got %d", typ, got.Type)
		}
	}
}

func TestResponseError(t *testing.T) {
	for _, err := range []Error{ErrorNone, ErrorIncorrectPassword, ErrorServerDoesntExist} {
		resp := Response{Type: ResponseTypeError, Error: err}
		buf := make([]byte, unsafe.Sizeof(Response{}))
		resp.Serialize(buf)
		got := DeserializeResponse(buf)
		if got.Error != err {
			t.Errorf("Error=%d round-trip failed: got %d", err, got.Error)
		}
		if got.Type != ResponseTypeError {
			t.Errorf("Type = %d, want %d", got.Type, ResponseTypeError)
		}
	}
}

func TestResponseJoinServerInfo(t *testing.T) {
	resp := Response{Type: ResponseTypeJoinServerInfo}
	copy(resp.Address[:], "10.20.30.40")
	resp.TotalList = 1

	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buf)

	got := DeserializeResponse(buf)

	if got.Type != ResponseTypeJoinServerInfo {
		t.Errorf("Type = %d, want %d", got.Type, ResponseTypeJoinServerInfo)
	}
	if string(addressTrim(got)) != "10.20.30.40" {
		t.Errorf("Address = %q, want %q", string(addressTrim(got)), "10.20.30.40")
	}
}

func TestResponseClientJoinInfo(t *testing.T) {
	resp := Response{Type: ResponseTypeClientJoinInfo}
	copy(resp.Address[:], "172.16.0.5")

	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buf)

	got := DeserializeResponse(buf)

	if got.Type != ResponseTypeClientJoinInfo {
		t.Errorf("Type = %d, want %d", got.Type, ResponseTypeClientJoinInfo)
	}
	if string(addressTrim(got)) != "172.16.0.5" {
		t.Errorf("Address = %q, want %q", string(addressTrim(got)), "172.16.0.5")
	}
}

func TestResponseBufferSize(t *testing.T) {
	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp := Response{}
	result := resp.Serialize(buf)

	if len(result) != int(unsafe.Sizeof(Response{})) {
		t.Errorf("Serialize returned buffer of size %d, want %d",
			len(result), int(unsafe.Sizeof(Response{})))
	}
}

func TestResponseEmptyEntries(t *testing.T) {
	resp := Response{Type: ResponseTypeServerList, TotalList: 0}
	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buf)

	got := DeserializeResponse(buf)
	if got.TotalList != 0 {
		t.Errorf("TotalList = %d, want 0", got.TotalList)
	}
	// All list entries should be zeroed
	for i := 0; i < serversPerResponse; i++ {
		if got.List[i].Id != 0 {
			t.Errorf("List[%d].Id should be 0, got %d", i, got.List[i].Id)
		}
	}
}

func TestResponseConstants(t *testing.T) {
	if serversPerResponse != 10 {
		t.Errorf("serversPerResponse = %d, want 10", serversPerResponse)
	}
	if addressMaxLen != 64 {
		t.Errorf("addressMaxLen = %d, want 64", addressMaxLen)
	}
}

func TestResponseAddressMaxLength(t *testing.T) {
	addr := "0123456789012345678901234567890123456789012345678901234567890123" // 64 chars
	resp := Response{Type: ResponseTypeJoinServerInfo}
	copy(resp.Address[:], addr)

	buf := make([]byte, unsafe.Sizeof(Response{}))
	resp.Serialize(buf)

	got := DeserializeResponse(buf)
	if string(got.Address[:len(addr)]) != addr {
		t.Errorf("64-char address round-trip failed: got %q", string(got.Address[:len(addr)]))
	}
}

func addressTrim(r Response) []byte {
	addr := r.Address[:]
	for i, b := range addr {
		if b == 0 {
			return addr[:i]
		}
	}
	return addr
}
