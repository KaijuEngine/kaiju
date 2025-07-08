/******************************************************************************/
/* master_server_request.go                                                   */
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
