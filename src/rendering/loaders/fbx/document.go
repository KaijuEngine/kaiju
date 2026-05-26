/******************************************************************************/
/* document.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const (
	BinaryHeader = "Kaydara FBX Binary  \x00\x1a\x00"
)

var (
	ErrASCIINotSupported = errors.New("ASCII FBX is not supported yet")
	ErrInvalidHeader     = errors.New("invalid FBX file")
)

type Document struct {
	Version uint32
	Nodes   []Node
}

type Node struct {
	Name          string
	Properties    []Property
	Children      []Node
	StartOffset   int64
	EndOffset     int64
	PropertyStart int64
	PropertyEnd   int64
}

type Property struct {
	Type   byte
	Offset int64
	Value  any
}

type ParseError struct {
	Offset int64
	Reason string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("fbx parse offset %d: %s", e.Offset, e.Reason)
}

func Parse(data []byte) (Document, error) {
	if isASCII(data) {
		return Document{}, ErrASCIINotSupported
	}
	if !bytes.HasPrefix(data, []byte(BinaryHeader)) {
		return Document{}, ErrInvalidHeader
	}
	versionOffset := len(BinaryHeader)
	if len(data) < versionOffset+4 {
		return Document{}, ErrInvalidHeader
	}
	doc := Document{
		Version: binary.LittleEndian.Uint32(data[versionOffset : versionOffset+4]),
	}
	reader := binaryReader{
		data:   data,
		cursor: versionOffset + 4,
		is64:   doc.Version >= 7500,
	}
	for reader.cursor < len(data) {
		node, ok, err := reader.readNode(len(data))
		if err != nil {
			return Document{}, err
		}
		if !ok {
			break
		}
		doc.Nodes = append(doc.Nodes, node)
	}
	return doc, nil
}

func isASCII(data []byte) bool {
	const sniffBytes = 256
	data = bytes.TrimLeft(data, " \t\r\n")
	if len(data) == 0 {
		return false
	}
	check := data
	if len(check) > sniffBytes {
		check = check[:sniffBytes]
	}
	return bytes.HasPrefix(check, []byte("Kaydara FBX ASCII")) ||
		(bytes.HasPrefix(check, []byte(";")) && bytes.Contains(check, []byte("FBX")))
}

type binaryReader struct {
	data   []byte
	cursor int
	is64   bool
}

func (r *binaryReader) readNode(limit int) (Node, bool, error) {
	start := r.cursor
	endOffset, err := r.readOffset(limit)
	if err != nil {
		return Node{}, false, err
	}
	if endOffset == 0 {
		return Node{}, false, nil
	}
	if endOffset > uint64(limit) {
		return Node{}, false, r.errAt(start, "record end offset is out of bounds")
	}
	if endOffset < uint64(r.cursor) {
		return Node{}, false, r.errAt(start, "record end offset is before the record header")
	}
	propCount, err := r.readOffset(limit)
	if err != nil {
		return Node{}, false, err
	}
	propLength, err := r.readOffset(limit)
	if err != nil {
		return Node{}, false, err
	}
	nameLength, err := r.readU8(limit)
	if err != nil {
		return Node{}, false, err
	}
	nameStart := r.cursor
	if err := r.need(int(nameLength), limit, "record name is out of bounds"); err != nil {
		return Node{}, false, err
	}
	name := string(r.data[nameStart : nameStart+int(nameLength)])
	r.cursor += int(nameLength)

	propStart := r.cursor
	propEnd64 := uint64(propStart) + propLength
	if propEnd64 > endOffset || propEnd64 > uint64(limit) {
		return Node{}, false, r.errAt(propStart, "property list length is out of bounds")
	}
	if propCount > uint64(int(^uint(0)>>1)) {
		return Node{}, false, r.errAt(propStart, "property count is too large")
	}
	props := make([]Property, 0, int(propCount))
	for i := uint64(0); i < propCount; i++ {
		prop, err := r.readProperty(int(propEnd64))
		if err != nil {
			return Node{}, false, err
		}
		props = append(props, prop)
	}
	if r.cursor != int(propEnd64) {
		return Node{}, false, r.errAt(r.cursor, "property list length was not fully consumed")
	}

	node := Node{
		Name:          name,
		Properties:    props,
		StartOffset:   int64(start),
		EndOffset:     int64(endOffset),
		PropertyStart: int64(propStart),
		PropertyEnd:   int64(propEnd64),
	}
	if r.cursor < int(endOffset) {
		sentinelLength := r.sentinelLength()
		childLimit := int(endOffset) - sentinelLength
		if childLimit < r.cursor {
			return Node{}, false, r.errAt(r.cursor, "insufficient sentinel bytes at record end")
		}
		for r.cursor < childLimit {
			child, ok, err := r.readNode(childLimit)
			if err != nil {
				return Node{}, false, err
			}
			if !ok {
				return Node{}, false, r.errAt(r.cursor, "unexpected null record before sentinel")
			}
			node.Children = append(node.Children, child)
		}
		if err := r.readSentinel(int(endOffset)); err != nil {
			return Node{}, false, err
		}
	}
	if r.cursor != int(endOffset) {
		return Node{}, false, r.errAt(r.cursor, "record end offset was not reached")
	}
	return node, true, nil
}

func (r *binaryReader) readProperty(limit int) (Property, error) {
	offset := r.cursor
	propType, err := r.readU8(limit)
	if err != nil {
		return Property{}, err
	}
	prop := Property{Type: propType, Offset: int64(offset)}
	switch propType {
	case 'Y':
		v, err := r.readU16(limit)
		prop.Value = int16(v)
		return prop, err
	case 'C':
		v, err := r.readU8(limit)
		prop.Value = v != 0
		return prop, err
	case 'I':
		v, err := r.readU32(limit)
		prop.Value = int32(v)
		return prop, err
	case 'F':
		v, err := r.readU32(limit)
		prop.Value = math.Float32frombits(v)
		return prop, err
	case 'D':
		v, err := r.readU64(limit)
		prop.Value = math.Float64frombits(v)
		return prop, err
	case 'L':
		v, err := r.readU64(limit)
		prop.Value = int64(v)
		return prop, err
	case 'S':
		value, err := r.readByteBlock(limit)
		prop.Value = string(value)
		return prop, err
	case 'R':
		value, err := r.readByteBlock(limit)
		prop.Value = value
		return prop, err
	case 'f', 'd', 'l', 'i', 'b', 'c':
		value, err := r.readArray(propType, limit)
		prop.Value = value
		return prop, err
	default:
		return Property{}, r.errAt(offset, fmt.Sprintf("unknown property type %q", propType))
	}
}

func (r *binaryReader) readByteBlock(limit int) ([]byte, error) {
	length, err := r.readU32(limit)
	if err != nil {
		return nil, err
	}
	if uint64(length) > uint64(int(^uint(0)>>1)) {
		return nil, r.errAt(r.cursor, "byte block length is too large")
	}
	if err := r.need(int(length), limit, "byte block is out of bounds"); err != nil {
		return nil, err
	}
	value := append([]byte(nil), r.data[r.cursor:r.cursor+int(length)]...)
	r.cursor += int(length)
	return value, nil
}

func (r *binaryReader) readArray(propType byte, limit int) (any, error) {
	offset := r.cursor - 1
	length, err := r.readU32(limit)
	if err != nil {
		return nil, err
	}
	encoding, err := r.readU32(limit)
	if err != nil {
		return nil, err
	}
	payloadLength, err := r.readU32(limit)
	if err != nil {
		return nil, err
	}
	stride := arrayElementSize(propType)
	if stride == 0 {
		return nil, r.errAt(offset, fmt.Sprintf("unknown array property type %q", propType))
	}
	if uint64(length)*uint64(stride) > uint64(int(^uint(0)>>1)) {
		return nil, r.errAt(offset, "array length is too large")
	}
	decodedLength := int(length) * stride
	if uint64(payloadLength) > uint64(int(^uint(0)>>1)) {
		return nil, r.errAt(r.cursor, "array payload length is too large")
	}
	if err := r.need(int(payloadLength), limit, "array payload is out of bounds"); err != nil {
		return nil, err
	}
	payload := r.data[r.cursor : r.cursor+int(payloadLength)]
	r.cursor += int(payloadLength)

	var bytes []byte
	switch encoding {
	case 0:
		if int(payloadLength) != decodedLength {
			return nil, r.errAt(offset, "raw array payload length does not match element count")
		}
		bytes = payload
	case 1:
		bytes, err = inflateZlib(payload, decodedLength)
		if err != nil {
			return nil, r.errAt(offset, "bad compressed array payload: "+err.Error())
		}
	default:
		return nil, r.errAt(offset, fmt.Sprintf("unknown array encoding %d", encoding))
	}
	if len(bytes) != decodedLength {
		return nil, r.errAt(offset, "array payload length does not match element count")
	}
	return decodeArray(propType, bytes), nil
}

func inflateZlib(payload []byte, decodedLength int) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var out bytes.Buffer
	if decodedLength > 0 {
		out.Grow(decodedLength)
	}
	if _, err := io.Copy(&out, zr); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func decodeArray(propType byte, data []byte) any {
	switch propType {
	case 'f':
		out := make([]float32, len(data)/4)
		for i := range out {
			out[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
		}
		return out
	case 'd':
		out := make([]float64, len(data)/8)
		for i := range out {
			out[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[i*8:]))
		}
		return out
	case 'l':
		out := make([]int64, len(data)/8)
		for i := range out {
			out[i] = int64(binary.LittleEndian.Uint64(data[i*8:]))
		}
		return out
	case 'i':
		out := make([]int32, len(data)/4)
		for i := range out {
			out[i] = int32(binary.LittleEndian.Uint32(data[i*4:]))
		}
		return out
	case 'b':
		out := make([]bool, len(data))
		for i := range out {
			out[i] = data[i] != 0
		}
		return out
	case 'c':
		return append([]byte(nil), data...)
	default:
		return nil
	}
}

func arrayElementSize(propType byte) int {
	switch propType {
	case 'f', 'i':
		return 4
	case 'd', 'l':
		return 8
	case 'b', 'c':
		return 1
	default:
		return 0
	}
}

func (r *binaryReader) readSentinel(end int) error {
	if err := r.need(end-r.cursor, end, "sentinel is out of bounds"); err != nil {
		return err
	}
	for i := r.cursor; i < end; i++ {
		if r.data[i] != 0 {
			return r.errAt(i, "bad nested record sentinel")
		}
	}
	r.cursor = end
	return nil
}

func (r *binaryReader) sentinelLength() int {
	if r.is64 {
		return 25
	}
	return 13
}

func (r *binaryReader) readOffset(limit int) (uint64, error) {
	if r.is64 {
		return r.readU64(limit)
	}
	v, err := r.readU32(limit)
	return uint64(v), err
}

func (r *binaryReader) readU8(limit int) (uint8, error) {
	if err := r.need(1, limit, "unexpected end of file"); err != nil {
		return 0, err
	}
	v := r.data[r.cursor]
	r.cursor++
	return v, nil
}

func (r *binaryReader) readU16(limit int) (uint16, error) {
	if err := r.need(2, limit, "unexpected end of file"); err != nil {
		return 0, err
	}
	v := binary.LittleEndian.Uint16(r.data[r.cursor:])
	r.cursor += 2
	return v, nil
}

func (r *binaryReader) readU32(limit int) (uint32, error) {
	if err := r.need(4, limit, "unexpected end of file"); err != nil {
		return 0, err
	}
	v := binary.LittleEndian.Uint32(r.data[r.cursor:])
	r.cursor += 4
	return v, nil
}

func (r *binaryReader) readU64(limit int) (uint64, error) {
	if err := r.need(8, limit, "unexpected end of file"); err != nil {
		return 0, err
	}
	v := binary.LittleEndian.Uint64(r.data[r.cursor:])
	r.cursor += 8
	return v, nil
}

func (r *binaryReader) need(count int, limit int, reason string) error {
	if count < 0 || r.cursor+count < r.cursor || r.cursor+count > limit || r.cursor+count > len(r.data) {
		return r.errAt(r.cursor, reason)
	}
	return nil
}

func (r *binaryReader) errAt(offset int, reason string) error {
	return ParseError{Offset: int64(offset), Reason: reason}
}
