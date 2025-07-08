package master_server

import (
	"encoding/binary"
	"unsafe"
)

type MasterServerRequestType = uint8

const (
	RequestTypeRegisterServer = MasterServerRequestType(iota)
	RequestTypeUnregisterServer
	RequestTypePing
	RequestTypeServerList
	RequestTypeJoinServer
)

type Request struct {
	Game           [gameKeySize]byte
	Name           [gameNameSize]byte
	Password       [MaxPasswordSize]byte
	ServerId       uint64
	MaxPlayers     uint16
	CurrentPlayers uint16
	Type           MasterServerRequestType
}

func (r *Request) Serialize(buffer []byte) {
	offset := copy(buffer, r.Game[:])
	offset += copy(buffer[offset:], r.Name[:])
	offset += copy(buffer[offset:], r.Password[:])
	binary.LittleEndian.PutUint16(buffer[offset:], r.MaxPlayers)
	offset += int(unsafe.Sizeof(r.MaxPlayers))
	binary.LittleEndian.PutUint16(buffer[offset:], r.CurrentPlayers)
	offset += int(unsafe.Sizeof(r.CurrentPlayers))
	buffer[offset] = r.Type
}

func DeserializeRequest(buffer []byte) Request {
	r := Request{}
	offset := copy(r.Game[:], buffer)
	offset += copy(r.Name[:], buffer[offset:])
	offset += copy(r.Password[:], buffer[offset:])
	r.MaxPlayers = binary.LittleEndian.Uint16(buffer[offset:])
	offset += int(unsafe.Sizeof(r.MaxPlayers))
	r.CurrentPlayers = binary.LittleEndian.Uint16(buffer[offset:])
	offset += int(unsafe.Sizeof(r.CurrentPlayers))
	r.Type = buffer[offset]
	return r
}
