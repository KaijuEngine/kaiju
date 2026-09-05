/******************************************************************************/
/* memory_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"unsafe"
)

// ---------------------------------------------------------------------------
// FindFirstZeroInByteArray / Memcpy
// ---------------------------------------------------------------------------

func TestFindFirstZeroInByteArray(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0}
	idx := FindFirstZeroInByteArray(data)
	if idx != 8 {
		t.Errorf("FindFirstZeroInByteArray(data) = %d, expected 8", idx)
	}
}

func TestFindFirstZeroInByteArray_NoZero(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	idx := FindFirstZeroInByteArray(data)
	if idx != 0 {
		t.Errorf("FindFirstZeroInByteArray(data) = %d, expected 0 (no zero found)", idx)
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
	Memcpy(unsafe.Pointer(&data[0]), unsafe.Pointer(&fromData[0]), uint64(bufferSize))
	for i := 0; i < bufferSize; i++ {
		if data[i] != byte(i+9) {
			t.Errorf("data[%d] = %d, expected %d", i, data[i], i+9)
		}
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

// byteLimitWriter only allows a fixed number of bytes to be written before
// failing. It is used to trigger errors in the middle of multi-write helpers.
type byteLimitWriter struct {
	remaining int
}

func (w *byteLimitWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("write limit reached")
	}
	if len(p) > w.remaining {
		n := w.remaining
		w.remaining = 0
		return n, errors.New("write limit reached")
	}
	w.remaining -= len(p)
	return len(p), nil
}

// memTestStruct is a fixed-size struct used to validate the unsafe byte-array
// conversion helpers. It intentionally contains no pointers or strings so the
// memory layout is plain value bytes.
type memTestStruct struct {
	A uint32
	B uint16
	C uint8
}

func int32Bytes(v int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return buf
}

// ---------------------------------------------------------------------------
// BinaryWrite / BinaryRead basics
// ---------------------------------------------------------------------------

func TestBinaryWrite_Int32(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWrite(&buf, int32(-12345)); err != nil {
		t.Fatalf("BinaryWrite returned error: %v", err)
	}
	got := buf.Bytes()
	if len(got) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(got))
	}
	want := int32(-12345)
	if binary.LittleEndian.Uint32(got) != uint32(want) {
		t.Fatalf("unexpected bytes: %v", got)
	}
}

func TestBinaryWrite_Error(t *testing.T) {
	if err := BinaryWrite(failingWriter{}, int32(1)); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestBinaryWriteInt(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteInt(&buf, 424242); err != nil {
		t.Fatalf("BinaryWriteInt returned error: %v", err)
	}
	if got := binary.LittleEndian.Uint32(buf.Bytes()); got != 424242 {
		t.Fatalf("got %d, expected 424242", got)
	}
}

func TestBinaryWriteInt_Error(t *testing.T) {
	if err := BinaryWriteInt(failingWriter{}, 1); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestBinaryRead(t *testing.T) {
	var v int32
	if err := BinaryRead(bytes.NewReader(int32Bytes(-777)), &v); err != nil {
		t.Fatalf("BinaryRead returned error: %v", err)
	}
	if v != -777 {
		t.Fatalf("got %d, expected -777", v)
	}
}

func TestBinaryRead_Error(t *testing.T) {
	var v int32
	if err := BinaryRead(failingReader{}, &v); err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestBinaryReadInt(t *testing.T) {
	v, err := BinaryReadInt(bytes.NewReader(int32Bytes(1234)))
	if err != nil {
		t.Fatalf("BinaryReadInt returned error: %v", err)
	}
	if v != 1234 {
		t.Fatalf("got %d, expected 1234", v)
	}
}

func TestBinaryReadInt_Error(t *testing.T) {
	if _, err := BinaryReadInt(failingReader{}); err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestBinaryReadLen(t *testing.T) {
	v, err := BinaryReadLen(bytes.NewReader(int32Bytes(7)))
	if err != nil {
		t.Fatalf("BinaryReadLen returned error: %v", err)
	}
	if v != 7 {
		t.Fatalf("got %d, expected 7", v)
	}
}

func TestBinaryReadLen_Error(t *testing.T) {
	if _, err := BinaryReadLen(bytes.NewReader(nil)); err == nil {
		t.Fatal("expected EOF error from empty reader")
	}
}

func TestBinaryReadVar(t *testing.T) {
	v, err := BinaryReadVar[int32](bytes.NewReader(int32Bytes(55)))
	if err != nil {
		t.Fatalf("BinaryReadVar returned error: %v", err)
	}
	if v != 55 {
		t.Fatalf("got %d, expected 55", v)
	}
}

func TestBinaryReadVar_Slice(t *testing.T) {
	v, err := BinaryReadVar[uint16](bytes.NewReader([]byte{0x34, 0x12}))
	if err != nil {
		t.Fatalf("BinaryReadVar returned error: %v", err)
	}
	if v != 0x1234 {
		t.Fatalf("got %x, expected 0x1234", v)
	}
}

func TestBinaryReadVar_Error(t *testing.T) {
	if _, err := BinaryReadVar[int32](failingReader{}); err == nil {
		t.Fatal("expected error from failing reader")
	}
}

// ---------------------------------------------------------------------------
// BinaryWriteSlice / BinaryReadVarSlice
// ---------------------------------------------------------------------------

func TestBinaryWriteSliceLen(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteSliceLen(&buf, []int32{1, 2, 3}); err != nil {
		t.Fatalf("BinaryWriteSliceLen returned error: %v", err)
	}
	if binary.LittleEndian.Uint32(buf.Bytes()) != 3 {
		t.Fatalf("expected length 3, got %v", buf.Bytes())
	}
}

func TestBinaryWriteSliceLen_Error(t *testing.T) {
	if err := BinaryWriteSliceLen(failingWriter{}, []int32{1}); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestBinaryWriteSlice(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteSlice(&buf, []int32{10, 20, 30}); err != nil {
		t.Fatalf("BinaryWriteSlice returned error: %v", err)
	}
	got, err := BinaryReadVarSlice[int32](bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if len(got) != 3 || got[0] != 10 || got[1] != 20 || got[2] != 30 {
		t.Fatalf("got %v, expected [10 20 30]", got)
	}
}

func TestBinaryWriteSlice_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteSlice(&buf, []int32{}); err != nil {
		t.Fatalf("BinaryWriteSlice returned error: %v", err)
	}
	if len(buf.Bytes()) != 4 || binary.LittleEndian.Uint32(buf.Bytes()) != 0 {
		t.Fatalf("expected only zero length, got %v", buf.Bytes())
	}
}

func TestBinaryWriteSlice_Error(t *testing.T) {
	if err := BinaryWriteSlice(failingWriter{}, []int32{1}); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestBinaryReadVarSlice_ReadError(t *testing.T) {
	if _, err := BinaryReadVarSlice[int32](failingReader{}); err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestBinaryReadVarSlice_NegativeLength(t *testing.T) {
	_, err := BinaryReadVarSlice[int32](bytes.NewReader(int32Bytes(-1)))
	if err == nil || err.Error() != "negative length read" {
		t.Fatalf("expected negative length error, got %v", err)
	}
}

func TestBinaryReadVarSlice_Empty(t *testing.T) {
	got, err := BinaryReadVarSlice[int32](bytes.NewReader(int32Bytes(0)))
	if err != nil {
		t.Fatalf("BinaryReadVarSlice returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestBinaryReadVarSlice_Truncated(t *testing.T) {
	data := append(int32Bytes(3), 0x01, 0x02) // length says 3 elements but only 2 bytes
	if _, err := BinaryReadVarSlice[int32](bytes.NewReader(data)); err == nil {
		t.Fatal("expected error for truncated slice data")
	}
}

// ---------------------------------------------------------------------------
// BinaryWriteString / BinaryReadString
// ---------------------------------------------------------------------------

func TestBinaryWriteString(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteString(&buf, "hello"); err != nil {
		t.Fatalf("BinaryWriteString returned error: %v", err)
	}
	b := buf.Bytes()
	if binary.LittleEndian.Uint32(b) != 5 || string(b[4:]) != "hello" {
		t.Fatalf("unexpected bytes: %v", b)
	}
}

func TestBinaryWriteString_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteString(&buf, ""); err != nil {
		t.Fatalf("BinaryWriteString returned error: %v", err)
	}
	if len(buf.Bytes()) != 4 || binary.LittleEndian.Uint32(buf.Bytes()) != 0 {
		t.Fatalf("expected zero length only, got %v", buf.Bytes())
	}
}

func TestBinaryWriteString_Error(t *testing.T) {
	if err := BinaryWriteString(failingWriter{}, "x"); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestBinaryWriteString_Truncated(t *testing.T) {
	// Length prefix fits (4 bytes) but the string body would exceed the limit.
	w := &byteLimitWriter{remaining: 4}
	if err := BinaryWriteString(w, "abc"); err == nil {
		t.Fatal("expected error when body exceeds write limit")
	}
}

func TestBinaryReadString(t *testing.T) {
	data := append(int32Bytes(5), []byte("world")...)
	s, err := BinaryReadString(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("BinaryReadString returned error: %v", err)
	}
	if s != "world" {
		t.Fatalf("got %q, expected %q", s, "world")
	}
}

func TestBinaryReadString_Empty(t *testing.T) {
	s, err := BinaryReadString(bytes.NewReader(int32Bytes(0)))
	if err != nil {
		t.Fatalf("BinaryReadString returned error: %v", err)
	}
	if s != "" {
		t.Fatalf("got %q, expected empty", s)
	}
}

func TestBinaryReadString_LengthError(t *testing.T) {
	if _, err := BinaryReadString(bytes.NewReader(nil)); err == nil {
		t.Fatal("expected EOF error from empty reader")
	}
}

func TestBinaryReadString_NegativeLength(t *testing.T) {
	_, err := BinaryReadString(bytes.NewReader(int32Bytes(-5)))
	if err == nil || err.Error() != "negative length read" {
		t.Fatalf("expected negative length error, got %v", err)
	}
}

func TestBinaryReadString_Truncated(t *testing.T) {
	data := append(int32Bytes(10), []byte("abc")...) // length says 10, only 3 present
	if _, err := BinaryReadString(bytes.NewReader(data)); err == nil {
		t.Fatal("expected error for truncated string data")
	}
}

// ---------------------------------------------------------------------------
// BinaryWriteStringSlice / BinaryReadStringSlice
// ---------------------------------------------------------------------------

func TestBinaryWriteStringSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []string{"one", "", "three"}
	if err := BinaryWriteStringSlice(&buf, data); err != nil {
		t.Fatalf("BinaryWriteStringSlice returned error: %v", err)
	}
	got, err := BinaryReadStringSlice(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if len(got) != 3 || got[0] != "one" || got[1] != "" || got[2] != "three" {
		t.Fatalf("got %v, expected %v", got, data)
	}
}

func TestBinaryWriteStringSlice_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteStringSlice(&buf, []string{}); err != nil {
		t.Fatalf("BinaryWriteStringSlice returned error: %v", err)
	}
	if binary.LittleEndian.Uint32(buf.Bytes()) != 0 {
		t.Fatalf("expected zero length, got %v", buf.Bytes())
	}
}

func TestBinaryWriteStringSlice_LengthError(t *testing.T) {
	if err := BinaryWriteStringSlice(failingWriter{}, []string{"a"}); err == nil {
		t.Fatal("expected error when length prefix cannot be written")
	}
}

func TestBinaryWriteStringSlice_StringError(t *testing.T) {
	// Only the length prefix fits; the first string body fails.
	w := &byteLimitWriter{remaining: 4}
	if err := BinaryWriteStringSlice(w, []string{"abcdef"}); err == nil {
		t.Fatal("expected error when a string body cannot be written")
	}
}

func TestBinaryReadStringSlice_LengthError(t *testing.T) {
	if _, err := BinaryReadStringSlice(bytes.NewReader(nil)); err == nil {
		t.Fatal("expected EOF error from empty reader")
	}
}

func TestBinaryReadStringSlice_StringError(t *testing.T) {
	// Length says 2 entries, first is fine, second is truncated.
	data := append(int32Bytes(2), int32Bytes(3)...)
	data = append(data, 'h', 'i')
	data = append(data, int32Bytes(5)...)
	data = append(data, 'a', 'b') // only 2 of the 5 declared bytes present
	if _, err := BinaryReadStringSlice(bytes.NewReader(data)); err == nil {
		t.Fatal("expected error for truncated string entry")
	}
}

// ---------------------------------------------------------------------------
// BinaryWriteMapLen / BinaryWriteMap
// ---------------------------------------------------------------------------

func TestBinaryWriteMapLen(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteMapLen(&buf, map[int32]int32{1: 10, 2: 20}); err != nil {
		t.Fatalf("BinaryWriteMapLen returned error: %v", err)
	}
	if binary.LittleEndian.Uint32(buf.Bytes()) != 2 {
		t.Fatalf("expected length 2, got %v", buf.Bytes())
	}
}

func TestBinaryWriteMapLen_Error(t *testing.T) {
	if err := BinaryWriteMapLen(failingWriter{}, map[int32]int32{1: 10}); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestBinaryWriteMap(t *testing.T) {
	var buf bytes.Buffer
	input := map[int32]int32{1: 100, 2: 200}
	if err := BinaryWriteMap(&buf, input); err != nil {
		t.Fatalf("BinaryWriteMap returned error: %v", err)
	}
	r := bytes.NewReader(buf.Bytes())
	l, err := BinaryReadLen(r)
	if err != nil {
		t.Fatalf("failed to read length: %v", err)
	}
	if l != int32(len(input)) {
		t.Fatalf("length = %d, expected %d", l, len(input))
	}
	got := make(map[int32]int32)
	for range l {
		k, err := BinaryReadVar[int32](r)
		if err != nil {
			t.Fatalf("failed to read key: %v", err)
		}
		v, err := BinaryReadVar[int32](r)
		if err != nil {
			t.Fatalf("failed to read value: %v", err)
		}
		got[k] = v
	}
	if len(got) != len(input) {
		t.Fatalf("got %v, expected %v", got, input)
	}
	for k, v := range input {
		if got[k] != v {
			t.Fatalf("got %v, expected %v", got, input)
		}
	}
}

func TestBinaryWriteMap_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := BinaryWriteMap(&buf, map[int32]int32{}); err != nil {
		t.Fatalf("BinaryWriteMap returned error: %v", err)
	}
	if len(buf.Bytes()) != 4 || binary.LittleEndian.Uint32(buf.Bytes()) != 0 {
		t.Fatalf("expected zero length only, got %v", buf.Bytes())
	}
}

func TestBinaryWriteMap_LengthError(t *testing.T) {
	if err := BinaryWriteMap(failingWriter{}, map[int32]int32{1: 10}); err == nil {
		t.Fatal("expected error when length prefix cannot be written")
	}
}

func TestBinaryWriteMap_KeyError(t *testing.T) {
	// Length prefix fits, but the first key write fails.
	w := &byteLimitWriter{remaining: 4}
	if err := BinaryWriteMap(w, map[int32]int32{1: 10}); err == nil {
		t.Fatal("expected error when key cannot be written")
	}
}

func TestBinaryWriteMap_ValueError(t *testing.T) {
	// Length prefix (4) + one key (4) fit, then the value write fails.
	w := &byteLimitWriter{remaining: 8}
	if err := BinaryWriteMap(w, map[int32]int32{1: 10}); err == nil {
		t.Fatal("expected error when value cannot be written")
	}
}

// ---------------------------------------------------------------------------
// Unsafe byte-array conversions
// ---------------------------------------------------------------------------

func TestInterfaceUnderlyingPointer(t *testing.T) {
	value := int32(42)
	var iface any = &value
	ptr := InterfaceUnderlyingPointer(iface)
	if ptr != unsafe.Pointer(&value) {
		t.Fatalf("ptr = %p, expected %p", ptr, &value)
	}
	if got := *(*int32)(ptr); got != 42 {
		t.Fatalf("got %d, expected 42", got)
	}
}

func TestStructToByteArray(t *testing.T) {
	s := memTestStruct{A: 0x01020304, B: 0x0506, C: 0x07}
	b := StructToByteArray(s)
	if len(b) != int(unsafe.Sizeof(s)) {
		t.Fatalf("len = %d, expected %d", len(b), unsafe.Sizeof(s))
	}
	back := *(*memTestStruct)(unsafe.Pointer(&b[0]))
	if back != s {
		t.Fatalf("roundtrip mismatch: got %+v, expected %+v", back, s)
	}
	if b[0] != 0x04 || b[1] != 0x03 || b[2] != 0x02 || b[3] != 0x01 {
		t.Fatalf("expected little-endian field A, got %v", b[:4])
	}
}

func TestSizedStructToByteArray(t *testing.T) {
	s := memTestStruct{A: 0xDEADBEEF, B: 0x1234, C: 0x99}
	size := int(unsafe.Sizeof(s))
	b := SizedStructToByteArray(unsafe.Pointer(&s), size)
	if len(b) != size {
		t.Fatalf("len = %d, expected %d", len(b), size)
	}
	back := *(*memTestStruct)(unsafe.Pointer(&b[0]))
	if back != s {
		t.Fatalf("roundtrip mismatch: got %+v, expected %+v", back, s)
	}
}

func TestStructSliceToByteArray(t *testing.T) {
	slice := []memTestStruct{
		{A: 1, B: 2, C: 3},
		{A: 4, B: 5, C: 6},
		{A: 7, B: 8, C: 9},
	}
	b := StructSliceToByteArray(slice)
	expectedLen := int(unsafe.Sizeof(slice[0])) * len(slice)
	if len(b) != expectedLen {
		t.Fatalf("len = %d, expected %d", len(b), expectedLen)
	}
	for i, s := range slice {
		back := *(*memTestStruct)(unsafe.Pointer(&b[i*int(unsafe.Sizeof(s))]))
		if back != s {
			t.Fatalf("element %d mismatch: got %+v, expected %+v", i, back, s)
		}
	}
}

func TestConvertByteSliceType(t *testing.T) {
	orig := []uint32{0x11223344, 0xAABBCCDD, 0x00000007}
	bytesIn := StructSliceToByteArray(orig)
	conv := ConvertByteSliceType[uint32](bytesIn)
	if len(conv) != len(orig) {
		t.Fatalf("len = %d, expected %d", len(conv), len(orig))
	}
	for i := range orig {
		if conv[i] != orig[i] {
			t.Fatalf("element %d = %x, expected %x", i, conv[i], orig[i])
		}
	}
}

func TestConvertByteSliceType_Int16(t *testing.T) {
	orig := []int16{-1, 0, 1, 300}
	bytesIn := StructSliceToByteArray(orig)
	conv := ConvertByteSliceType[int16](bytesIn)
	if len(conv) != len(orig) {
		t.Fatalf("len = %d, expected %d", len(conv), len(orig))
	}
	for i := range orig {
		if conv[i] != orig[i] {
			t.Fatalf("element %d = %d, expected %d", i, conv[i], orig[i])
		}
	}
}
