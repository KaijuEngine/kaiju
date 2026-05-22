/******************************************************************************/
/* master_server_constants_test.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package master_server

import "testing"

func TestConstants(t *testing.T) {
	if MaxPasswordSize != 16 {
		t.Errorf("MaxPasswordSize = %d, want 16", MaxPasswordSize)
	}
	if gameKeySize != 32 {
		t.Errorf("gameKeySize = %d, want 32", gameKeySize)
	}
	if gameNameSize != 64 {
		t.Errorf("gameNameSize = %d, want 64", gameNameSize)
	}
	if serversPerResponse != 10 {
		t.Errorf("serversPerResponse = %d, want 10", serversPerResponse)
	}
	if addressMaxLen != 64 {
		t.Errorf("addressMaxLen = %d, want 64", addressMaxLen)
	}
}

func TestErrorConstants(t *testing.T) {
	if ErrorNone != 0 {
		t.Errorf("ErrorNone = %d, want 0", ErrorNone)
	}
	if ErrorIncorrectPassword != 1 {
		t.Errorf("ErrorIncorrectPassword = %d, want 1", ErrorIncorrectPassword)
	}
	if ErrorServerDoesntExist != 2 {
		t.Errorf("ErrorServerDoesntExist = %d, want 2", ErrorServerDoesntExist)
	}
}

func TestRequestTypeConstants(t *testing.T) {
	if RequestTypeRegisterServer != 0 {
		t.Errorf("RequestTypeRegisterServer = %d, want 0", RequestTypeRegisterServer)
	}
	if RequestTypeUnregisterServer != 1 {
		t.Errorf("RequestTypeUnregisterServer = %d, want 1", RequestTypeUnregisterServer)
	}
	if RequestTypePing != 2 {
		t.Errorf("RequestTypePing = %d, want 2", RequestTypePing)
	}
	if RequestTypeServerList != 3 {
		t.Errorf("RequestTypeServerList = %d, want 3", RequestTypeServerList)
	}
	if RequestTypeJoinServer != 4 {
		t.Errorf("RequestTypeJoinServer = %d, want 4", RequestTypeJoinServer)
	}
}

func TestResponseTypeConstants(t *testing.T) {
	if ResponseTypeConfirmRegister != 0 {
		t.Errorf("ResponseTypeConfirmRegister = %d, want 0", ResponseTypeConfirmRegister)
	}
	if ResponseTypeServerList != 1 {
		t.Errorf("ResponseTypeServerList = %d, want 1", ResponseTypeServerList)
	}
	if ResponseTypeJoinServerInfo != 2 {
		t.Errorf("ResponseTypeJoinServerInfo = %d, want 2", ResponseTypeJoinServerInfo)
	}
	if ResponseTypeClientJoinInfo != 3 {
		t.Errorf("ResponseTypeClientJoinInfo = %d, want 3", ResponseTypeClientJoinInfo)
	}
	if ResponseTypeError != 4 {
		t.Errorf("ResponseTypeError = %d, want 4", ResponseTypeError)
	}
}

func TestResponseSize(t *testing.T) {
	// Verify Response struct size is consistent
	expectedSize := 1 + // Type
		serversPerResponse*(gameNameSize+8+2+2) + // List
		addressMaxLen + // Address
		4 + // TotalList
		1 // Error
	if int(responseStructSize()) != expectedSize {
		t.Errorf("Response struct size = %d, calculated = %d", responseStructSize(), expectedSize)
	}
}

func responseStructSize() int {
	return 1 + // Type byte
		serversPerResponse*(gameNameSize+8+2+2) + // each ResponseServerList: name(64) + id(8) + max(2) + cur(2)
		addressMaxLen + // Address byte array
		4 + // TotalList uint32
		1 // Error byte
}
