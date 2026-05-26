/******************************************************************************/
/* document_test.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

type testNode struct {
	name        string
	properties  [][]byte
	children    []testNode
	forceNested bool
}

func TestParseHeaderVersion7400(t *testing.T) {
	doc, err := Parse(testFBXFile(7400, testNullRecord(false)))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if doc.Version != 7400 {
		t.Fatalf("Version = %d, want 7400", doc.Version)
	}
	if len(doc.Nodes) != 0 {
		t.Fatalf("len(Nodes) = %d, want 0", len(doc.Nodes))
	}
}

func TestParseHeaderVersion7500Uses64BitRecordHeaders(t *testing.T) {
	node := testNode{name: "Root", properties: [][]byte{propString("wide")}}
	doc, err := Parse(testFBXFile(7500, append(node.build(7500, len(BinaryHeader)+4), testNullRecord(true)...)))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if doc.Version != 7500 {
		t.Fatalf("Version = %d, want 7500", doc.Version)
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Name != "Root" {
		t.Fatalf("Nodes = %#v, want one Root node", doc.Nodes)
	}
	if got := doc.Nodes[0].Properties[0].Value; got != "wide" {
		t.Fatalf("property = %#v, want %q", got, "wide")
	}
}

func TestParseScalarProperties(t *testing.T) {
	node := testNode{name: "Scalars", properties: [][]byte{
		propInt16(-1234),
		propBool(true),
		propInt32(-56789),
		propFloat32(12.5),
		propFloat64(-98.25),
		propInt64(-1234567890123),
		propString("hello"),
		propRaw([]byte{0x01, 0x02, 0x03}),
	}}
	doc, err := Parse(testFBXFileWithNodes(7400, node))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	props := doc.Nodes[0].Properties
	checkValue(t, props[0].Value, int16(-1234))
	checkValue(t, props[1].Value, true)
	checkValue(t, props[2].Value, int32(-56789))
	checkValue(t, props[3].Value, float32(12.5))
	checkValue(t, props[4].Value, float64(-98.25))
	checkValue(t, props[5].Value, int64(-1234567890123))
	checkValue(t, props[6].Value, "hello")
	checkValue(t, props[7].Value, []byte{0x01, 0x02, 0x03})
	for i, p := range props {
		if p.Offset <= 0 {
			t.Fatalf("property %d offset = %d, want source offset", i, p.Offset)
		}
	}
}

func TestParseArrayProperties(t *testing.T) {
	node := testNode{name: "Arrays", properties: [][]byte{
		propArrayRaw('f', floats32Bytes(1.5, -2.25)),
		propArrayRaw('d', floats64Bytes(3.25, -4.5)),
		propArrayRaw('l', int64sBytes(-5, 6)),
		propArrayRaw('i', int32sBytes(-7, 8)),
		propArrayRaw('b', []byte{1, 0, 1}),
		propArrayRaw('c', []byte("abc")),
		propArrayZlib('f', floats32Bytes(9.5, 10.5)),
	}}
	doc, err := Parse(testFBXFileWithNodes(7400, node))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	props := doc.Nodes[0].Properties
	checkValue(t, props[0].Value, []float32{1.5, -2.25})
	checkValue(t, props[1].Value, []float64{3.25, -4.5})
	checkValue(t, props[2].Value, []int64{-5, 6})
	checkValue(t, props[3].Value, []int32{-7, 8})
	checkValue(t, props[4].Value, []bool{true, false, true})
	checkValue(t, props[5].Value, []byte("abc"))
	checkValue(t, props[6].Value, []float32{9.5, 10.5})
}

func TestParseNestedNodesAndSentinel(t *testing.T) {
	parent := testNode{
		name:        "Parent",
		forceNested: true,
		children: []testNode{
			{name: "Child", properties: [][]byte{propInt32(42)}},
		},
	}
	doc, err := Parse(testFBXFileWithNodes(7400, parent))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(doc.Nodes[0].Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(doc.Nodes[0].Children))
	}
	child := doc.Nodes[0].Children[0]
	if child.Name != "Child" || child.Properties[0].Value != int32(42) {
		t.Fatalf("child = %#v, want Child with int32 property", child)
	}
}

func TestParseMalformedBounds(t *testing.T) {
	body := make([]byte, 13)
	binary.LittleEndian.PutUint32(body[0:4], uint32(len(BinaryHeader)+4+len(body)+100))
	body[12] = 0
	_, err := Parse(testFBXFile(7400, body))
	if err == nil {
		t.Fatal("Parse returned nil error, want malformed bounds error")
	}
	if !strings.Contains(err.Error(), "offset") || !strings.Contains(err.Error(), "out of bounds") {
		t.Fatalf("error = %q, want offset and out of bounds", err.Error())
	}
}

func TestParseUnknownTypeCode(t *testing.T) {
	node := testNode{name: "Bad", properties: [][]byte{{'Z'}}}
	_, err := Parse(testFBXFileWithNodes(7400, node))
	if err == nil {
		t.Fatal("Parse returned nil error, want unknown type error")
	}
	if !strings.Contains(err.Error(), "offset") || !strings.Contains(err.Error(), "unknown property type") {
		t.Fatalf("error = %q, want offset and unknown property type", err.Error())
	}
}

func TestParseBadSentinel(t *testing.T) {
	node := testNode{name: "Parent", forceNested: true}
	data := testFBXFileWithNodes(7400, node)
	data[len(data)-13-1] = 1
	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse returned nil error, want bad sentinel error")
	}
	if !strings.Contains(err.Error(), "offset") || !strings.Contains(err.Error(), "sentinel") {
		t.Fatalf("error = %q, want offset and sentinel", err.Error())
	}
}

func TestParseBadCompressedPayload(t *testing.T) {
	prop := []byte{'f'}
	prop = binary.LittleEndian.AppendUint32(prop, 2)
	prop = binary.LittleEndian.AppendUint32(prop, 1)
	prop = binary.LittleEndian.AppendUint32(prop, 4)
	prop = append(prop, 0x01, 0x02, 0x03, 0x04)
	node := testNode{name: "BadCompressed", properties: [][]byte{prop}}
	_, err := Parse(testFBXFileWithNodes(7400, node))
	if err == nil {
		t.Fatal("Parse returned nil error, want bad compressed payload error")
	}
	if !strings.Contains(err.Error(), "offset") || !strings.Contains(err.Error(), "bad compressed array payload") {
		t.Fatalf("error = %q, want offset and bad compressed array payload", err.Error())
	}
}

func TestParseMonkeyFixtureBinaryTree(t *testing.T) {
	data, err := os.ReadFile("../../../editor/editor_embedded_content/editor_content/meshes/monkey.fbx")
	if err != nil {
		t.Skipf("monkey fixture not available: %v", err)
	}
	doc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(monkey.fbx) returned error: %v", err)
	}
	if doc.Version != 7400 {
		t.Fatalf("Version = %d, want 7400", doc.Version)
	}
	if !testHasNode(doc.Nodes, "Geometry") {
		t.Fatalf("monkey.fbx did not contain a Geometry node")
	}
	if !testHasNode(doc.Nodes, "Model") {
		t.Fatalf("monkey.fbx did not contain a Model node")
	}
}

func checkValue(t *testing.T, got any, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("value = %#v, want %#v", got, want)
	}
}

func testFBXFileWithNodes(version uint32, nodes ...testNode) []byte {
	body := []byte{}
	offset := len(BinaryHeader) + 4
	for _, node := range nodes {
		encoded := node.build(version, offset)
		body = append(body, encoded...)
		offset += len(encoded)
	}
	body = append(body, testNullRecord(version >= 7500)...)
	return testFBXFile(version, body)
}

func testFBXFile(version uint32, body []byte) []byte {
	data := append([]byte(BinaryHeader), 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(data[len(BinaryHeader):], version)
	data = append(data, body...)
	return data
}

func (n testNode) build(version uint32, start int) []byte {
	is64 := version >= 7500
	headerLength := 13
	sentinelLength := 13
	if is64 {
		headerLength = 25
		sentinelLength = 25
	}
	props := bytes.Join(n.properties, nil)
	children := []byte{}
	childOffset := start + headerLength + len(n.name) + len(props)
	for _, child := range n.children {
		encoded := child.build(version, childOffset)
		children = append(children, encoded...)
		childOffset += len(encoded)
	}
	if len(children) > 0 || n.forceNested {
		children = append(children, make([]byte, sentinelLength)...)
	}
	end := start + headerLength + len(n.name) + len(props) + len(children)
	out := make([]byte, 0, end-start)
	if is64 {
		out = binary.LittleEndian.AppendUint64(out, uint64(end))
		out = binary.LittleEndian.AppendUint64(out, uint64(len(n.properties)))
		out = binary.LittleEndian.AppendUint64(out, uint64(len(props)))
	} else {
		out = binary.LittleEndian.AppendUint32(out, uint32(end))
		out = binary.LittleEndian.AppendUint32(out, uint32(len(n.properties)))
		out = binary.LittleEndian.AppendUint32(out, uint32(len(props)))
	}
	out = append(out, byte(len(n.name)))
	out = append(out, n.name...)
	out = append(out, props...)
	out = append(out, children...)
	return out
}

func testNullRecord(is64 bool) []byte {
	if is64 {
		return make([]byte, 25)
	}
	return make([]byte, 13)
}

func propInt16(v int16) []byte {
	return binary.LittleEndian.AppendUint16([]byte{'Y'}, uint16(v))
}

func propBool(v bool) []byte {
	if v {
		return []byte{'C', 1}
	}
	return []byte{'C', 0}
}

func propInt32(v int32) []byte {
	return binary.LittleEndian.AppendUint32([]byte{'I'}, uint32(v))
}

func propFloat32(v float32) []byte {
	return binary.LittleEndian.AppendUint32([]byte{'F'}, math.Float32bits(v))
}

func propFloat64(v float64) []byte {
	return binary.LittleEndian.AppendUint64([]byte{'D'}, math.Float64bits(v))
}

func propInt64(v int64) []byte {
	return binary.LittleEndian.AppendUint64([]byte{'L'}, uint64(v))
}

func propString(v string) []byte {
	out := binary.LittleEndian.AppendUint32([]byte{'S'}, uint32(len(v)))
	return append(out, v...)
}

func propRaw(v []byte) []byte {
	out := binary.LittleEndian.AppendUint32([]byte{'R'}, uint32(len(v)))
	return append(out, v...)
}

func propArrayRaw(propType byte, data []byte) []byte {
	out := []byte{propType}
	out = binary.LittleEndian.AppendUint32(out, uint32(len(data)/testArrayStride(propType)))
	out = binary.LittleEndian.AppendUint32(out, 0)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(data)))
	return append(out, data...)
}

func propArrayZlib(propType byte, data []byte) []byte {
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(data); err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	out := []byte{propType}
	out = binary.LittleEndian.AppendUint32(out, uint32(len(data)/testArrayStride(propType)))
	out = binary.LittleEndian.AppendUint32(out, 1)
	out = binary.LittleEndian.AppendUint32(out, uint32(compressed.Len()))
	return append(out, compressed.Bytes()...)
}

func floats32Bytes(values ...float32) []byte {
	out := []byte{}
	for _, v := range values {
		out = binary.LittleEndian.AppendUint32(out, math.Float32bits(v))
	}
	return out
}

func floats64Bytes(values ...float64) []byte {
	out := []byte{}
	for _, v := range values {
		out = binary.LittleEndian.AppendUint64(out, math.Float64bits(v))
	}
	return out
}

func int32sBytes(values ...int32) []byte {
	out := []byte{}
	for _, v := range values {
		out = binary.LittleEndian.AppendUint32(out, uint32(v))
	}
	return out
}

func int64sBytes(values ...int64) []byte {
	out := []byte{}
	for _, v := range values {
		out = binary.LittleEndian.AppendUint64(out, uint64(v))
	}
	return out
}

func testArrayStride(propType byte) int {
	switch propType {
	case 'f', 'i':
		return 4
	case 'd', 'l':
		return 8
	case 'b', 'c':
		return 1
	default:
		panic("unknown test array type")
	}
}

func testHasNode(nodes []Node, name string) bool {
	for _, node := range nodes {
		if node.Name == name || testHasNode(node.Children, name) {
			return true
		}
	}
	return false
}
