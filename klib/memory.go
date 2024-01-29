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

func BinaryWrite(w io.Writer, data any) {
	binary.Write(w, binary.LittleEndian, data)
}

func BinaryWriteSliceLen[T any](w io.Writer, data []T) {
	BinaryWrite(w, int32(len(data)))
}

func BinaryWriteSlice[T any](w io.Writer, data []T) {
	BinaryWriteSliceLen[T](w, data)
	if len(data) > 0 {
		BinaryWrite(w, data)
	}
}

func BinaryWriteMapLen[K comparable, V any](w io.Writer, data map[K]V) {
	BinaryWrite(w, int32(len(data)))
}

func BinaryWriteMap[K comparable, V any](w io.Writer, data map[K]V) {
	BinaryWriteMapLen[K](w, data)
	for k, v := range data {
		BinaryWrite(w, k)
		BinaryWrite(w, v)
	}
}

func BinaryRead(r io.Reader, data any) {
	binary.Read(r, binary.LittleEndian, data)
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
	binary.Read(r, binary.LittleEndian, &length)
	if length < 0 {
		return nil, errors.New("negative length read")
	}
	if length > 0 {
		data := make([]T, length)
		binary.Read(r, binary.LittleEndian, &data)
		return data, nil
	} else {
		return []T{}, nil
	}
}

func BinaryWriteString(w io.Writer, str string) {
	length := int32(len(str))
	binary.Write(w, binary.LittleEndian, length)
	if length > 0 {
		binary.Write(w, binary.LittleEndian, []byte(str))
	}
}

func BinaryReadString(r io.Reader) (string, error) {
	var length int32
	binary.Read(r, binary.LittleEndian, &length)
	if length < 0 {
		return "", errors.New("negative length read")
	}
	if length > 0 {
		buff := make([]byte, length)
		binary.Read(r, binary.LittleEndian, &buff)
		return string(buff), nil
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

func Memcpy(dst unsafe.Pointer, src unsafe.Pointer, size int) {
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
	Memcpy(unsafe.Pointer(&res[0]), unsafe.Pointer(&slice[0]), count)
	return res
}
