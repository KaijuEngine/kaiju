/******************************************************************************/
/* gltf_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package loaders

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

type gltfTestBuilder struct {
	data        []byte
	bufferViews []map[string]any
	accessors   []map[string]any
}

type gltfTestVertex struct {
	position matrix.Vec3
	normal   matrix.Vec3
	uv       matrix.Vec2
}

func TestGLTFWithOptionsParallelMatchesSerial(t *testing.T) {
	db := testGLTFDatabase(t, "scene.gltf", "scene.bin", 3)
	serial, err := GLTFWithOptions("scene.gltf", db, GLTFLoadOptions{Workers: 1})
	if err != nil {
		t.Fatalf("serial GLTFWithOptions returned error: %v", err)
	}
	parallel, err := GLTFWithOptions("scene.gltf", db, GLTFLoadOptions{Workers: 4})
	if err != nil {
		t.Fatalf("parallel GLTFWithOptions returned error: %v", err)
	}
	assertGLTFResultsEqual(t, serial, parallel)
}

func TestGLTFAccessorsAndPrimitiveData(t *testing.T) {
	db := testGLTFDatabase(t, "scene.gltf", "scene.bin", 3)
	res, err := GLTFWithOptions("scene.gltf", db, GLTFLoadOptions{Workers: 4})
	if err != nil {
		t.Fatalf("GLTFWithOptions returned error: %v", err)
	}
	if len(res.Meshes) != 3 {
		t.Fatalf("mesh count = %d, want 3", len(res.Meshes))
	}
	if got, want := res.Meshes[0].MeshName, "scene.gltf/MeshA"; got != want {
		t.Fatalf("mesh 0 key = %q, want %q", got, want)
	}
	if got, want := res.Meshes[1].MeshName, "scene.gltf/MeshA_2"; got != want {
		t.Fatalf("mesh 1 key = %q, want %q", got, want)
	}
	if got, want := res.Meshes[2].MeshName, "scene.gltf/MeshB"; got != want {
		t.Fatalf("mesh 2 key = %q, want %q", got, want)
	}
	if got, want := res.Meshes[0].Verts[1].Position, (matrix.Vec3{1, 2, 0}); got != want {
		t.Fatalf("interleaved position = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[0].Verts[2].UV0, (matrix.Vec2{0, 1}); got != want {
		t.Fatalf("strided UV = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[0].Indexes, []uint32{0, 1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("uint8 indices = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[1].Indexes, []uint32{0, 2, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("uint16 indices = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[2].Indexes, []uint32{2, 0, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("uint32 indices = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[0].Verts[0].Color, matrix.ColorRed(); got != want {
		t.Fatalf("primitive 0 color = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[1].Verts[0].Color, matrix.ColorGreen(); got != want {
		t.Fatalf("primitive 1 color = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[1].Verts[2].Position, (matrix.Vec3{0, 1, 1}); got != want {
		t.Fatalf("byte-offset position = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[1].Verts[2].MorphTarget, (matrix.Vec3{0.3, 0.4, 0.5}); got != want {
		t.Fatalf("morph target = %#v, want %#v", got, want)
	}
	if got, want := res.Meshes[0].Textures["baseColor"], "embedded_0_baseColor"; got != want {
		t.Fatalf("baseColor texture = %q, want %q", got, want)
	}
	if got, want := res.TextureBytes["embedded_0_baseColor"], []byte{1, 2, 3, 4}; !bytes.Equal(got, want) {
		t.Fatalf("embedded texture bytes = %#v, want %#v", got, want)
	}
	if !res.Nodes[0].IsAnimated || !res.Nodes[1].IsAnimated {
		t.Fatalf("expected animated child and parent nodes, got root=%v child=%v", res.Nodes[0].IsAnimated, res.Nodes[1].IsAnimated)
	}
	if res.Nodes[2].IsAnimated {
		t.Fatal("unanimated sibling node was marked animated")
	}
}

func BenchmarkGLTFSinglePrimitive(b *testing.B) {
	db := testGLTFMultiPrimitiveDatabase(b, "bench.gltf", "bench.bin", 1, 50_000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := GLTF("bench.gltf", db); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGLTFMultiPrimitive(b *testing.B) {
	db := testGLTFMultiPrimitiveDatabase(b, "bench.gltf", "bench.bin", 8, 10_000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := GLTF("bench.gltf", db); err != nil {
			b.Fatal(err)
		}
	}
}

func assertGLTFResultsEqual(t *testing.T, a, b any) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("GLTF results differ\nserial=%#v\nparallel=%#v", a, b)
	}
}

func testGLTFDatabase(t testing.TB, gltfPath, binPath string, firstPrimitiveVerts int) *assets.MockDatabase {
	t.Helper()
	jsonData, binData := testGLTFFixture(t, binPath, firstPrimitiveVerts)
	return assets.NewMockDB(map[string][]byte{
		gltfPath: []byte(jsonData),
		binPath:  binData,
	})
}

func testGLTFMultiPrimitiveDatabase(t testing.TB, gltfPath, binPath string, primitiveCount, vertexCount int) *assets.MockDatabase {
	t.Helper()
	builder := &gltfTestBuilder{}
	meshPrimitives := make([]map[string]any, primitiveCount)
	for i := range meshPrimitives {
		pos, nml, uv, idx := builder.addTightPrimitiveAccessors(testGLTFGeneratedVertices(vertexCount, matrix.Float(i)), testU32Indices(vertexCount))
		meshPrimitives[i] = map[string]any{
			"attributes": map[string]any{
				"POSITION":   pos,
				"NORMAL":     nml,
				"TEXCOORD_0": uv,
			},
			"indices": idx,
			"mode":    4,
		}
	}
	doc := map[string]any{
		"asset":       map[string]any{"version": "2.0"},
		"buffers":     []map[string]any{{"uri": binPath, "byteLength": len(builder.data)}},
		"bufferViews": builder.bufferViews,
		"accessors":   builder.accessors,
		"meshes":      []map[string]any{{"name": "BenchMesh", "primitives": meshPrimitives}},
		"nodes":       []map[string]any{{"name": "BenchNode", "mesh": 0}},
	}
	jsonData, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return assets.NewMockDB(map[string][]byte{
		gltfPath: jsonData,
		binPath:  builder.data,
	})
}

func testGLTFFixture(t testing.TB, binPath string, firstPrimitiveVerts int) (string, []byte) {
	t.Helper()
	builder := &gltfTestBuilder{}
	builder.addBufferView([]byte{0, 0, 0, 0}, 0)
	pos0, nml0, uv0, idx0 := builder.addInterleavedPrimitiveAccessors(testGLTFGeneratedVertices(firstPrimitiveVerts, 0), testU8Indices(firstPrimitiveVerts))
	pos1, nml1, uv1, idx1 := builder.addOffsetPrimitiveAccessors([]gltfTestVertex{
		{matrix.Vec3{0, 0, 1}, matrix.Vec3{0, 0, 1}, matrix.Vec2{0, 0}},
		{matrix.Vec3{1, 0, 1}, matrix.Vec3{0, 0, 1}, matrix.Vec2{1, 0}},
		{matrix.Vec3{0, 1, 1}, matrix.Vec3{0, 0, 1}, matrix.Vec2{0, 1}},
	}, []uint16{0, 2, 1})
	morph := builder.addMorphAccessor([]matrix.Vec3{{0.1, 0.2, 0.3}, {0.2, 0.3, 0.4}, {0.3, 0.4, 0.5}})
	pos2, nml2, uv2, idx2 := builder.addTightPrimitiveAccessors([]gltfTestVertex{
		{matrix.Vec3{0, 0, 2}, matrix.Vec3{0, 1, 0}, matrix.Vec2{0, 0}},
		{matrix.Vec3{1, 0, 2}, matrix.Vec3{0, 1, 0}, matrix.Vec2{1, 0}},
		{matrix.Vec3{0, 1, 2}, matrix.Vec3{0, 1, 0}, matrix.Vec2{0, 1}},
	}, []uint32{2, 0, 1})
	animInput := builder.addFloatAccessor(f32Bytes(0, 1), 0, 5126, 2, "SCALAR")
	animOutput := builder.addFloatAccessor(f32Bytes(0, 0, 0, 1, 2, 3), 0, 5126, 2, "VEC3")
	embedded := []byte{1, 2, 3, 4}
	doc := map[string]any{
		"asset":       map[string]any{"version": "2.0"},
		"buffers":     []map[string]any{{"uri": binPath, "byteLength": len(builder.data)}},
		"bufferViews": builder.bufferViews,
		"accessors":   builder.accessors,
		"materials": []map[string]any{
			{"pbrMetallicRoughness": map[string]any{
				"baseColorFactor":  []float32{1, 0, 0, 1},
				"baseColorTexture": map[string]any{"index": 0},
			}},
			{"pbrMetallicRoughness": map[string]any{
				"baseColorFactor": []float32{0, 1, 0, 1},
			}},
		},
		"textures": []map[string]any{{"source": 0}},
		"images": []map[string]any{{
			"uri": "data:image/png;base64," + base64.StdEncoding.EncodeToString(embedded),
		}},
		"meshes": []map[string]any{
			{
				"name": "MeshA",
				"primitives": []map[string]any{
					{
						"attributes": map[string]any{
							"POSITION":   pos0,
							"NORMAL":     nml0,
							"TEXCOORD_0": uv0,
						},
						"indices":  idx0,
						"material": 0,
						"mode":     4,
					},
					{
						"attributes": map[string]any{
							"POSITION":   pos1,
							"NORMAL":     nml1,
							"TEXCOORD_0": uv1,
						},
						"indices":  idx1,
						"material": 1,
						"targets":  []map[string]any{{"POSITION": morph}},
						"mode":     4,
					},
				},
			},
			{
				"name": "MeshB",
				"primitives": []map[string]any{{
					"attributes": map[string]any{
						"POSITION":   pos2,
						"NORMAL":     nml2,
						"TEXCOORD_0": uv2,
					},
					"indices": idx2,
					"mode":    4,
				}},
			},
		},
		"nodes": []map[string]any{
			{"name": "Root", "children": []int{1, 2}},
			{"name": "NodeA", "mesh": 0},
			{"name": "NodeB", "mesh": 1},
		},
		"animations": []map[string]any{{
			"name": "Move",
			"channels": []map[string]any{{
				"sampler": 0,
				"target":  map[string]any{"node": 1, "path": "translation"},
			}},
			"samplers": []map[string]any{{
				"input":         animInput,
				"output":        animOutput,
				"interpolation": "LINEAR",
			}},
		}},
	}
	jsonData, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(jsonData), builder.data
}

func (b *gltfTestBuilder) addInterleavedPrimitiveAccessors(vertices []gltfTestVertex, indices []byte) (int, int, int, int) {
	view := b.addBufferView(interleavedVertexBytes(vertices), 32)
	pos := b.addAccessor(view, 0, 5126, len(vertices), "VEC3")
	nml := b.addAccessor(view, 12, 5126, len(vertices), "VEC3")
	uv := b.addAccessor(view, 24, 5126, len(vertices), "VEC2")
	idxView := b.addBufferView(indices, 0)
	idx := b.addAccessor(idxView, 0, 5121, len(indices), "SCALAR")
	return pos, nml, uv, idx
}

func (b *gltfTestBuilder) addOffsetPrimitiveAccessors(vertices []gltfTestVertex, indices []uint16) (int, int, int, int) {
	posBytes := append([]byte{9, 9, 9, 9}, vec3Bytes(testPositions(vertices))...)
	posView := b.addBufferView(posBytes, 0)
	nmlView := b.addBufferView(vec3Bytes(testNormals(vertices)), 0)
	uvView := b.addBufferView(vec2Bytes(testUVs(vertices)), 0)
	idxView := b.addBufferView(u16Bytes(indices...), 0)
	pos := b.addAccessor(posView, 4, 5126, len(vertices), "VEC3")
	nml := b.addAccessor(nmlView, 0, 5126, len(vertices), "VEC3")
	uv := b.addAccessor(uvView, 0, 5126, len(vertices), "VEC2")
	idx := b.addAccessor(idxView, 0, 5123, len(indices), "SCALAR")
	return pos, nml, uv, idx
}

func (b *gltfTestBuilder) addTightPrimitiveAccessors(vertices []gltfTestVertex, indices []uint32) (int, int, int, int) {
	posView := b.addBufferView(vec3Bytes(testPositions(vertices)), 0)
	nmlView := b.addBufferView(vec3Bytes(testNormals(vertices)), 0)
	uvView := b.addBufferView(vec2Bytes(testUVs(vertices)), 0)
	idxView := b.addBufferView(u32Bytes(indices...), 0)
	pos := b.addAccessor(posView, 0, 5126, len(vertices), "VEC3")
	nml := b.addAccessor(nmlView, 0, 5126, len(vertices), "VEC3")
	uv := b.addAccessor(uvView, 0, 5126, len(vertices), "VEC2")
	idx := b.addAccessor(idxView, 0, 5125, len(indices), "SCALAR")
	return pos, nml, uv, idx
}

func (b *gltfTestBuilder) addMorphAccessor(points []matrix.Vec3) int {
	view := b.addBufferView(append([]byte{7, 7, 7, 7}, vec3Bytes(points)...), 0)
	return b.addAccessor(view, 4, 5126, len(points), "VEC3")
}

func (b *gltfTestBuilder) addFloatAccessor(data []byte, byteOffset, componentType, count int, accessorType string) int {
	view := b.addBufferView(data, 0)
	return b.addAccessor(view, byteOffset, componentType, count, accessorType)
}

func (b *gltfTestBuilder) addBufferView(data []byte, byteStride int) int {
	offset := len(b.data)
	b.data = append(b.data, data...)
	view := map[string]any{
		"buffer":     0,
		"byteOffset": offset,
		"byteLength": len(data),
	}
	if byteStride > 0 {
		view["byteStride"] = byteStride
	}
	b.bufferViews = append(b.bufferViews, view)
	return len(b.bufferViews) - 1
}

func (b *gltfTestBuilder) addAccessor(bufferView, byteOffset, componentType, count int, accessorType string) int {
	b.accessors = append(b.accessors, map[string]any{
		"bufferView":    bufferView,
		"byteOffset":    byteOffset,
		"componentType": componentType,
		"count":         count,
		"type":          accessorType,
	})
	return len(b.accessors) - 1
}

func testGLTFGeneratedVertices(count int, z matrix.Float) []gltfTestVertex {
	verts := make([]gltfTestVertex, count)
	for i := range verts {
		x := matrix.Float(i)
		verts[i] = gltfTestVertex{
			position: matrix.Vec3{x, x + 1, z},
			normal:   matrix.Vec3{0, 0, 1},
			uv:       matrix.Vec2{matrix.Float(i % 2), matrix.Float((i / 2) % 2)},
		}
	}
	return verts
}

func testU8Indices(count int) []byte {
	out := make([]byte, count)
	for i := range out {
		out[i] = byte(i % count)
	}
	return out
}

func testU32Indices(count int) []uint32 {
	out := make([]uint32, count)
	for i := range out {
		out[i] = uint32(i % count)
	}
	return out
}

func interleavedVertexBytes(vertices []gltfTestVertex) []byte {
	out := make([]byte, 0, len(vertices)*32)
	for i := range vertices {
		out = append(out, vec3Bytes([]matrix.Vec3{vertices[i].position})...)
		out = append(out, vec3Bytes([]matrix.Vec3{vertices[i].normal})...)
		out = append(out, vec2Bytes([]matrix.Vec2{vertices[i].uv})...)
	}
	return out
}

func testPositions(vertices []gltfTestVertex) []matrix.Vec3 {
	out := make([]matrix.Vec3, len(vertices))
	for i := range vertices {
		out[i] = vertices[i].position
	}
	return out
}

func testNormals(vertices []gltfTestVertex) []matrix.Vec3 {
	out := make([]matrix.Vec3, len(vertices))
	for i := range vertices {
		out[i] = vertices[i].normal
	}
	return out
}

func testUVs(vertices []gltfTestVertex) []matrix.Vec2 {
	out := make([]matrix.Vec2, len(vertices))
	for i := range vertices {
		out[i] = vertices[i].uv
	}
	return out
}

func vec3Bytes(values []matrix.Vec3) []byte {
	out := make([]byte, 0, len(values)*12)
	for i := range values {
		out = append(out, f32Bytes(float32(values[i].X()), float32(values[i].Y()), float32(values[i].Z()))...)
	}
	return out
}

func vec2Bytes(values []matrix.Vec2) []byte {
	out := make([]byte, 0, len(values)*8)
	for i := range values {
		out = append(out, f32Bytes(float32(values[i].X()), float32(values[i].Y()))...)
	}
	return out
}

func f32Bytes(values ...float32) []byte {
	out := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(v))
	}
	return out
}

func u16Bytes(values ...uint16) []byte {
	out := make([]byte, len(values)*2)
	for i, v := range values {
		binary.LittleEndian.PutUint16(out[i*2:], v)
	}
	return out
}

func u32Bytes(values ...uint32) []byte {
	out := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(out[i*4:], v)
	}
	return out
}
