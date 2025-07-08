package master_server

import (
	"encoding/binary"
	"unsafe"
)

type MasterServerResponseType = uint8

const (
	ResponseTypeConfirmRegister = MasterServerResponseType(iota)
	ResponseTypeServerList
	ResponseTypeJoinServerInfo
	ResponseTypeClientJoinInfo
	ResponseTypeError

	serversPerResponse = 10
	addressMaxLen      = 64
)

type Response struct {
	Type      MasterServerResponseType
	List      [serversPerResponse]ResponseServerList
	Address   [addressMaxLen]byte
	TotalList uint32
	Error     uint8
}

type ResponseServerList struct {
	Name           [gameNameSize]byte
	Id             uint64
	MaxPlayers     uint16
	CurrentPlayers uint16
}

func (r Response) Serialize(buffer []byte) []byte {
	offset := 0
	buffer[offset] = r.Type
	offset++
	for i := range r.List {
		offset += copy(buffer[offset:], r.List[i].Name[:])
		binary.LittleEndian.PutUint64(buffer[offset:], r.List[i].Id)
		offset += int(unsafe.Sizeof(r.List[i].Id))
		binary.LittleEndian.PutUint16(buffer[offset:], r.List[i].MaxPlayers)
		offset += int(unsafe.Sizeof(r.List[i].MaxPlayers))
		binary.LittleEndian.PutUint16(buffer[offset:], r.List[i].CurrentPlayers)
		offset += int(unsafe.Sizeof(r.List[i].CurrentPlayers))
	}
	offset += copy(buffer[offset:], r.Address[:])
	binary.LittleEndian.PutUint32(buffer[offset:], r.TotalList)
	offset += int(unsafe.Sizeof(r.TotalList))
	buffer[offset] = r.Error
	return buffer
}

func DeserializeResponse(buffer []byte) Response {
	r := Response{}
	offset := 0
	r.Type = buffer[offset]
	offset++
	for i := range r.List {
		offset += copy(r.List[i].Name[:], buffer[offset:])
		r.List[i].Id = binary.LittleEndian.Uint64(buffer[offset:])
		offset += int(unsafe.Sizeof(r.List[i].Id))
		r.List[i].MaxPlayers = binary.LittleEndian.Uint16(buffer[offset:])
		offset += int(unsafe.Sizeof(r.List[i].MaxPlayers))
		r.List[i].CurrentPlayers = binary.LittleEndian.Uint16(buffer[offset:])
		offset += int(unsafe.Sizeof(r.List[i].CurrentPlayers))
	}
	offset += copy(r.Address[:], buffer[offset:])
	r.TotalList = binary.LittleEndian.Uint32(buffer[offset:])
	offset += int(unsafe.Sizeof(r.TotalList))
	r.Error = buffer[offset]
	return r
}
