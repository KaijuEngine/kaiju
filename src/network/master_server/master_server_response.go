/******************************************************************************/
/* master_server_response.go                                                  */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
