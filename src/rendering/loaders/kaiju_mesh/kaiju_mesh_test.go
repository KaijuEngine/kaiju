/******************************************************************************/
/* kaiju_mesh_test.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"bytes"
	"encoding/gob"
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestKaijuMeshSerializePreservesBVH(t *testing.T) {
	km := KaijuMesh{
		Name: "triangle",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}},
			{Position: matrix.Vec3{1, 0, 0}},
			{Position: matrix.Vec3{0, 1, 0}},
		},
		Indexes: []uint32{0, 1, 2},
	}
	km.EnsureBVH()
	if km.BVH == nil {
		t.Fatal("expected mesh import data to include a BVH archive")
	}
	data, err := km.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.BVH == nil {
		t.Fatal("expected serialized mesh to preserve the BVH archive")
	}
	bvh := loaded.GenerateBVH(nil, nil, "hit")
	target, _, ok := bvh.RayIntersect(graviton.Ray{
		Origin:    matrix.Vec3{0.25, 0.25, 1},
		Direction: matrix.Vec3{0, 0, -1},
	}, 2)
	if !ok {
		t.Fatal("expected restored BVH to intersect the triangle")
	}
	if target != "hit" {
		t.Fatalf("expected restored BVH data to be hydrated, got %v", target)
	}
}

func TestKaijuMeshDeserializeLegacyGob(t *testing.T) {
	km := KaijuMesh{
		Name: "legacy-gob-triangle",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}},
			{Position: matrix.Vec3{1, 0, 0}},
			{Position: matrix.Vec3{0, 1, 0}},
		},
		Indexes: []uint32{0, 1, 2},
	}
	km.EnsureBVH()
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(km); err != nil {
		t.Fatal(err)
	}
	loaded, err := Deserialize(data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != km.Name {
		t.Fatalf("expected mesh name %q, got %q", km.Name, loaded.Name)
	}
	if len(loaded.Verts) != len(km.Verts) {
		t.Fatalf("expected %d verts, got %d", len(km.Verts), len(loaded.Verts))
	}
	if loaded.BVH == nil {
		t.Fatal("expected legacy gob fallback to preserve the BVH archive")
	}
}

func TestKaijuMeshGenerateBVHFallsBackWhenArchiveMissing(t *testing.T) {
	km := KaijuMesh{
		Name: "legacy-triangle",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}},
			{Position: matrix.Vec3{1, 0, 0}},
			{Position: matrix.Vec3{0, 1, 0}},
		},
		Indexes: []uint32{0, 1, 2},
	}
	if bvh := km.GenerateBVH(nil, nil, nil); bvh == nil {
		t.Fatal("expected legacy mesh data to generate a fallback BVH")
	}
	if km.BVH == nil {
		t.Fatal("expected fallback BVH generation to populate the mesh archive")
	}
}

func BenchmarkKaijuMeshDeserializeNative(b *testing.B) {
	km := benchmarkMesh()
	data, err := km.Serialize()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for range b.N {
		if _, err := Deserialize(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkKaijuMeshDeserializeLegacyGob(b *testing.B) {
	km := benchmarkMesh()
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(km); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(data.Len()))
	b.ResetTimer()
	for range b.N {
		if _, err := Deserialize(data.Bytes()); err != nil {
			b.Fatal(err)
		}
	}
}

func TestKaijuMeshLoadAndGenerateBVH(t *testing.T) {
	km := KaijuMesh{
		Name: "test-bvh",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}},
			{Position: matrix.Vec3{1, 0, 0}},
			{Position: matrix.Vec3{0, 1, 0}},
		},
		Indexes: []uint32{0, 1, 2},
	}
	data, err := km.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	bvh := loaded.GenerateBVH(nil, nil, "test-data")
	if bvh == nil {
		t.Fatal("expected BVH to be generated")
	}
	ray := graviton.Ray{
		Origin:    matrix.Vec3{0.25, 0.25, 1},
		Direction: matrix.Vec3{0, 0, -1},
	}
	target, _, ok := bvh.RayIntersect(ray, 2)
	if !ok {
		t.Fatal("expected BVH ray intersection to succeed")
	}
	if target != "test-data" {
		t.Fatalf("expected BVH data to be %q, got %v", "test-data", target)
	}
}

func BenchmarkKaijuMeshLoadAndGenerateBVH(b *testing.B) {
	km := benchmarkMesh()
	data, err := km.Serialize()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for range b.N {
		loaded, err := Deserialize(data)
		if err != nil {
			b.Fatal(err)
		}
		bvh := loaded.GenerateBVH(nil, nil, nil)
		if bvh == nil {
			b.Fatal("expected BVH to be generated")
		}
	}
}

func benchmarkMesh() KaijuMesh {
	const vertexCount = 50_000
	const indexCount = 150_000
	km := KaijuMesh{
		Name:    "benchmark",
		Verts:   make([]rendering.Vertex, vertexCount),
		Indexes: make([]uint32, indexCount),
	}
	for i := range km.Verts {
		v := matrix.Float(i)
		km.Verts[i] = rendering.Vertex{
			Position:     matrix.Vec3{v, v + 1, v + 2},
			Normal:       matrix.Vec3{0, 1, 0},
			Tangent:      matrix.Vec4{1, 0, 0, 1},
			UV0:          matrix.Vec2{v, v},
			Color:        matrix.ColorWhite(),
			JointIds:     matrix.Vec4i{1, 2, 3, 4},
			JointWeights: matrix.Vec4{0.25, 0.25, 0.25, 0.25},
			MorphTarget:  matrix.Vec3{v + 3, v + 4, v + 5},
		}
	}
	for i := range km.Indexes {
		km.Indexes[i] = uint32(i % vertexCount)
	}
	return km
}
