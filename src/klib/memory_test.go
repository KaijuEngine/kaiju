package klib

import (
	"testing"
	"unsafe"
)

func TestFindFirstZeroInByteArray(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0}
	idx := FindFirstZeroInByteArray(data)
	if idx != 8 {
		t.Errorf("FindFirstZeroInByteArray(data) = %d, expected %d", 8, idx)
	}
}

func TestUnsafeMemcpy(t *testing.T) {
	fromData := [8]byte{9, 10, 11, 12, 13, 14, 15, 16}
	bufferSize := 8
	data := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	for i := 0; i < bufferSize; i++ {
		if data[i] != byte(i+1) {
			t.Errorf("data[%d] = %d, expected %d", i, data[i], i+1)
		}
	}
	Memcpy(unsafe.Pointer(&data[0]), unsafe.Pointer(&fromData[0]), bufferSize)
	for i := 0; i < bufferSize; i++ {
		if data[i] != byte(i+9) {
			t.Errorf("data[%d] = %d, expected %d", i, data[i], i+9)
		}
	}
}
