package pod

import (
	"bytes"
	"kaiju/matrix"
	"reflect"
	"testing"
)

// SimpleTypes tests encoding and decoding of basic primitive types
func TestSimpleTypes(t *testing.T) {
	type SimpleStruct struct {
		IntVal    int32
		FloatVal  float32
		StringVal string
	}
	// Register the test structure
	Register(SimpleStruct{})
	original := SimpleStruct{
		IntVal:    42,
		FloatVal:  3.14,
		StringVal: "hello",
	}
	buf := bytes.NewBuffer([]byte{})
	encoder := NewEncoder(buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded SimpleStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded != original {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}

// AllPrimitiveTypes tests all supported primitive types
func TestAllPrimitiveTypes(t *testing.T) {
	type AllTypes struct {
		Int8Val       int8
		Int16Val      int16
		Int32Val      int32
		Int64Val      int64
		Uint8Val      uint8
		Uint16Val     uint16
		Uint32Val     uint32
		Uint64Val     uint64
		Float32Val    float32
		Float64Val    float64
		Complex64Val  complex64
		Complex128Val complex128
		RuneVal       rune
		StringVal     string
	}
	Register(AllTypes{})
	original := AllTypes{
		Int8Val:       -8,
		Int16Val:      -16,
		Int32Val:      -32,
		Int64Val:      -64,
		Uint8Val:      8,
		Uint16Val:     16,
		Uint32Val:     32,
		Uint64Val:     64,
		Float32Val:    1.23,
		Float64Val:    4.56,
		Complex64Val:  complex(1, 2),
		Complex128Val: complex(3, 4),
		RuneVal:       'A',
		StringVal:     "test",
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded AllTypes
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded != original {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}

// VectorTypes tests encoding/decoding of matrix types
func TestVectorTypes(t *testing.T) {
	type VectorStruct struct {
		Vec2Val  matrix.Vec2
		Vec3Val  matrix.Vec3
		Vec4Val  matrix.Vec4
		ColorVal matrix.Color
		QuatVal  matrix.Quaternion
	}
	Register(VectorStruct{})
	original := VectorStruct{
		Vec2Val:  matrix.Vec2{1.5, 2.5},
		Vec3Val:  matrix.Vec3{1.0, 2.0, 3.0},
		Vec4Val:  matrix.Vec4{1.0, 2.0, 3.0, 4.0},
		ColorVal: matrix.Color{0.1, 0.2, 0.3, 1.0},
		QuatVal:  matrix.Quaternion{0.0, 0.0, 0.7071, 0.7071},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded VectorStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded.Vec2Val != original.Vec2Val {
		t.Errorf("Vec2 mismatch: got %v, want %v", decoded.Vec2Val, original.Vec2Val)
	}
	if decoded.Vec3Val != original.Vec3Val {
		t.Errorf("Vec3 mismatch: got %v, want %v", decoded.Vec3Val, original.Vec3Val)
	}
	if decoded.Vec4Val != original.Vec4Val {
		t.Errorf("Vec4 mismatch: got %v, want %v", decoded.Vec4Val, original.Vec4Val)
	}
	if decoded.ColorVal != original.ColorVal {
		t.Errorf("Color mismatch: got %v, want %v", decoded.ColorVal, original.ColorVal)
	}
	if decoded.QuatVal != original.QuatVal {
		t.Errorf("Quaternion mismatch: got %v, want %v", decoded.QuatVal, original.QuatVal)
	}
}

// NestedStructures tests encoding/decoding of nested struct types
func TestNestedStructures(t *testing.T) {
	type InnerStruct struct {
		Value  int32
		Name   string
		Vector matrix.Vec3
	}
	type OuterStruct struct {
		Inner1 InnerStruct
		Inner2 InnerStruct
		Count  uint32
	}
	Register(InnerStruct{})
	Register(OuterStruct{})
	original := OuterStruct{
		Inner1: InnerStruct{
			Value:  10,
			Name:   "first",
			Vector: matrix.Vec3{1.0, 2.0, 3.0},
		},
		Inner2: InnerStruct{
			Value:  20,
			Name:   "second",
			Vector: matrix.Vec3{4.0, 5.0, 6.0},
		},
		Count: 2,
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded OuterStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded != original {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}

// ArrayTypes tests encoding/decoding of arrays
func TestArrayTypes(t *testing.T) {
	type ArrayStruct struct {
		IntArray    [4]int32
		StringArray [2]string
		VectorArray [3]matrix.Vec2
	}
	Register(ArrayStruct{})
	original := ArrayStruct{
		IntArray:    [4]int32{10, 20, 30, 40},
		StringArray: [2]string{"hello", "world"},
		VectorArray: [3]matrix.Vec2{
			{1.0, 2.0},
			{3.0, 4.0},
			{5.0, 6.0},
		},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded ArrayStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded != original {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}

// SliceTypes tests encoding/decoding of slices
func TestSliceTypes(t *testing.T) {
	type SliceStruct struct {
		IntSlice    []int32
		StringSlice []string
		VectorSlice []matrix.Vec3
	}
	Register(SliceStruct{})
	original := SliceStruct{
		IntSlice:    []int32{1, 2, 3, 4, 5},
		StringSlice: []string{"foo", "bar", "baz"},
		VectorSlice: []matrix.Vec3{
			{1.0, 2.0, 3.0},
			{4.0, 5.0, 6.0},
		},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded SliceStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if len(decoded.IntSlice) != len(original.IntSlice) ||
		len(decoded.StringSlice) != len(original.StringSlice) ||
		len(decoded.VectorSlice) != len(original.VectorSlice) {
		t.Errorf("slice length mismatch")
	}
	for i, v := range original.IntSlice {
		if decoded.IntSlice[i] != v {
			t.Errorf("IntSlice[%d] mismatch: got %v, want %v", i, decoded.IntSlice[i], v)
		}
	}
	for i, v := range original.StringSlice {
		if decoded.StringSlice[i] != v {
			t.Errorf("StringSlice[%d] mismatch: got %v, want %v", i, decoded.StringSlice[i], v)
		}
	}
	for i, v := range original.VectorSlice {
		if decoded.VectorSlice[i] != v {
			t.Errorf("VectorSlice[%d] mismatch: got %v, want %v", i, decoded.VectorSlice[i], v)
		}
	}
}

// EmptySlice tests encoding/decoding of empty slices
func TestEmptySlice(t *testing.T) {
	type SliceStruct struct {
		IntSlice []int32
	}
	Register(SliceStruct{})
	original := SliceStruct{
		IntSlice: []int32{},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded SliceStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if len(decoded.IntSlice) != 0 {
		t.Errorf("expected empty slice, got length %d", len(decoded.IntSlice))
	}
}

// NestedSlices tests encoding/decoding of slices of structs
func TestNestedSlices(t *testing.T) {
	type Item struct {
		ID   int32
		Name string
	}
	type Container struct {
		Items []Item
	}
	Register(Item{})
	Register(Container{})
	original := Container{
		Items: []Item{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
			{ID: 3, Name: "third"},
		},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded Container
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if len(decoded.Items) != len(original.Items) {
		t.Fatalf("slice length mismatch: got %d, want %d", len(decoded.Items), len(original.Items))
	}
	for i, item := range original.Items {
		if decoded.Items[i] != item {
			t.Errorf("Items[%d] mismatch: got %v, want %v", i, decoded.Items[i], item)
		}
	}
}

// ComplexStructure tests a complex structure with multiple field types
func TestComplexStructure(t *testing.T) {
	type Address struct {
		Street string
		City   string
		Postal int32
	}
	type Person struct {
		Name    string
		Age     uint8
		Height  float32
		Address Address
		Scores  []int32
		Tags    [3]string
	}
	Register(Address{})
	Register(Person{})
	original := Person{
		Name:   "Alice",
		Age:    30,
		Height: 5.8,
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
			Postal: 12345,
		},
		Scores: []int32{95, 87, 92},
		Tags:   [3]string{"engineer", "runner", "reader"},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded Person
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded.Name != original.Name || decoded.Age != original.Age || decoded.Height != original.Height {
		t.Errorf("basic fields mismatch")
	}
	if decoded.Address != original.Address {
		t.Errorf("Address mismatch: got %v, want %v", decoded.Address, original.Address)
	}
	if len(decoded.Scores) != len(original.Scores) {
		t.Fatalf("Scores length mismatch")
	}
	for i, s := range original.Scores {
		if decoded.Scores[i] != s {
			t.Errorf("Scores[%d] mismatch: got %d, want %d", i, decoded.Scores[i], s)
		}
	}
	if decoded.Tags != original.Tags {
		t.Errorf("Tags mismatch: got %v, want %v", decoded.Tags, original.Tags)
	}
}

// PointersAreSkipped tests that pointer fields are not encoded/decoded
func TestPointersAreSkipped(t *testing.T) {
	type WithPointer struct {
		Value   int32
		Pointer *int32
		Name    string
	}
	Register(WithPointer{})
	ptrVal := int32(999)
	original := WithPointer{
		Value:   42,
		Pointer: &ptrVal,
		Name:    "test",
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded WithPointer
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	// The pointer field should not be encoded, so it won't be decoded
	if decoded.Value != original.Value {
		t.Errorf("Value mismatch: got %d, want %d", decoded.Value, original.Value)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	// Pointer should remain nil after decoding (it's skipped)
	if decoded.Pointer != nil {
		t.Errorf("Pointer should be nil after decode, got %v", *decoded.Pointer)
	}
}

// InterfacesAreSkipped tests that interface fields are not encoded/decoded
func TestInterfacesAreSkipped(t *testing.T) {
	type WithInterface struct {
		Value     int32
		Interface interface{}
		Name      string
	}
	Register(WithInterface{})
	original := WithInterface{
		Value:     42,
		Interface: "some value", // This will not be encoded
		Name:      "test",
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded WithInterface
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	// The interface field should not be encoded, so it won't be decoded
	if decoded.Value != original.Value {
		t.Errorf("Value mismatch: got %d, want %d", decoded.Value, original.Value)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	// Interface should remain nil after decoding (it's skipped)
	if decoded.Interface != nil {
		t.Errorf("Interface should be nil after decode, got %v", decoded.Interface)
	}
}

// ArrayOfStructs tests encoding/decoding of arrays containing structs
func TestArrayOfStructs(t *testing.T) {
	type Point struct {
		X float32
		Y float32
	}
	type Polygon struct {
		Points [4]Point
	}
	Register(Point{})
	Register(Polygon{})
	original := Polygon{
		Points: [4]Point{
			{0.0, 0.0},
			{1.0, 0.0},
			{1.0, 1.0},
			{0.0, 1.0},
		},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded Polygon
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded != original {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}

// SliceOfVectors tests slices of vector types
func TestSliceOfVectors(t *testing.T) {
	type Path struct {
		Points []matrix.Vec2
	}
	Register(Path{})
	original := Path{
		Points: []matrix.Vec2{
			{0.0, 0.0},
			{1.0, 1.0},
			{2.0, 0.0},
		},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded Path
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if len(decoded.Points) != len(original.Points) {
		t.Fatalf("Points length mismatch")
	}
	for i, p := range original.Points {
		if decoded.Points[i] != p {
			t.Errorf("Points[%d] mismatch: got %v, want %v", i, decoded.Points[i], p)
		}
	}
}

// DeepNesting tests deeply nested structures
func TestDeepNesting(t *testing.T) {
	type Level3 struct {
		Value int32
	}
	type Level2 struct {
		L3   Level3
		Name string
	}
	type Level1 struct {
		L2   Level2
		Flag uint8
	}
	Register(Level3{})
	Register(Level2{})
	Register(Level1{})
	original := Level1{
		L2: Level2{
			L3: Level3{
				Value: 123,
			},
			Name: "nested",
		},
		Flag: 1,
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded Level1
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded != original {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}

// ZeroValues tests encoding/decoding of zero/empty values
func TestZeroValues(t *testing.T) {
	type ZeroStruct struct {
		IntVal    int32
		StringVal string
		VecVal    matrix.Vec3
		SliceVal  []int32
	}
	Register(ZeroStruct{})
	original := ZeroStruct{
		IntVal:    0,
		StringVal: "",
		VecVal:    matrix.Vec3{0, 0, 0},
		SliceVal:  []int32{},
	}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded ZeroStruct
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if decoded.IntVal != 0 {
		t.Errorf("IntVal should be 0, got %d", decoded.IntVal)
	}
	if decoded.StringVal != "" {
		t.Errorf("StringVal should be empty, got %q", decoded.StringVal)
	}
	if decoded.VecVal != (matrix.Vec3{0, 0, 0}) {
		t.Errorf("VecVal should be zero, got %v", decoded.VecVal)
	}
	if len(decoded.SliceVal) != 0 {
		t.Errorf("SliceVal should be empty, got length %d", len(decoded.SliceVal))
	}
}

// RecursiveType tests encoding/decoding of a recursive slice type
func TestRecursiveType(t *testing.T) {
	type Me struct {
		Inner []Me
	}
	Register(Me{})
	// Construct a nested recursive structure
	original := Me{Inner: []Me{{Inner: []Me{}}, {Inner: []Me{{Inner: []Me{}}}}}}
	buf := bytes.Buffer{}
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(original); err != nil {
		t.Fatalf("encoding failed: %v", err)
	}
	var decoded Me
	decoder := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("decoded value mismatch: got %v, want %v", decoded, original)
	}
}
