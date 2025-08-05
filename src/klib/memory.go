/******************************************************************************/
/* memory.go                                                                  */
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

package klib

import (
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
)

type Serializable interface {
	Serialize(stream io.Writer)
	Deserialize(stream io.Reader)
}

func BinaryWrite(w io.Writer, data any) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func BinaryWriteSliceLen[T any](w io.Writer, data []T) error {
	return BinaryWrite(w, int32(len(data)))
}

func BinaryWriteSlice[T any](w io.Writer, data []T) error {
	err := BinaryWriteSliceLen[T](w, data)
	if err == nil && len(data) > 0 {
		return BinaryWrite(w, data)
	}
	return err
}

func BinaryWriteMapLen[K comparable, V any](w io.Writer, data map[K]V) error {
	return BinaryWrite(w, int32(len(data)))
}

func BinaryWriteMap[K comparable, V any](w io.Writer, data map[K]V) error {
	if err := BinaryWriteMapLen(w, data); err != nil {
		return err
	}
	for k, v := range data {
		if err := BinaryWrite(w, k); err != nil {
			return err
		}
		if err := BinaryWrite(w, v); err != nil {
			return err
		}
	}
	return nil
}

func BinaryRead(r io.Reader, data any) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func BinaryReadLen(r io.Reader) (int32, error) {
	return BinaryReadVar[int32](r)
}

func BinaryReadVar[T any](r io.Reader) (T, error) {
	var data T
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func BinaryReadVarSlice[T any](r io.Reader) ([]T, error) {
	var length int32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, errors.New("negative length read")
	}
	if length > 0 {
		data := make([]T, length)
		err := binary.Read(r, binary.LittleEndian, &data)
		return data, err
	} else {
		return []T{}, nil
	}
}

func BinaryWriteString(w io.Writer, str string) error {
	length := int32(len(str))
	err := binary.Write(w, binary.LittleEndian, length)
	if err != nil && length > 0 {
		return binary.Write(w, binary.LittleEndian, []byte(str))
	}
	return err
}

func BinaryReadString(r io.Reader) (string, error) {
	var length int32
	err := binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", errors.New("negative length read")
	}
	if length > 0 {
		buff := make([]byte, length)
		err := binary.Read(r, binary.LittleEndian, &buff)
		return string(buff), err
	} else {
		return "", nil
	}
}

func FindFirstZeroInByteArray(arr []byte) int {
	end := 0
	for i, b := range arr {
		if b == 0 {
			end = i
			break
		}
	}
	return end
}

func Memcpy(dst unsafe.Pointer, src unsafe.Pointer, size uint64) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func InterfaceUnderlyingPointer[T any](s T) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + uintptr(8)))
}

func StructToByteArray[T any](s T) []byte {
	const m = 0x7fffffff
	size := unsafe.Sizeof(s)
	return (*[m]byte)(unsafe.Pointer(&s))[:size]
}

func SizedStructToByteArray(s unsafe.Pointer, size int) []byte {
	const m = 0x7fffffff
	tmp := unsafe.Slice((*byte)(s), size)
	//return (*[m]byte)(unsafe.Pointer(&s))[:size]
	return (*[m]byte)(unsafe.Pointer(&tmp[0]))[:size]
}

func StructSliceToByteArray[T any](s []T) []byte {
	const m = 0x7fffffff
	size := int(unsafe.Sizeof(s[0])) * len(s)
	return (*[m]byte)(unsafe.Pointer(&s[0]))[:size]
}

func ConvertByteSliceType[T any](slice []byte) []T {
	count := len(slice)
	res := make([]T, count/int(unsafe.Sizeof(*(*T)(nil))))
	Memcpy(unsafe.Pointer(&res[0]), unsafe.Pointer(&slice[0]), uint64(count))
	return res
}
