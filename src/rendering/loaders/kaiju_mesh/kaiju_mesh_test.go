/******************************************************************************/
/* kaiju_mesh_test.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"bytes"
	"encoding/gob"
	"slices"
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	meshloaders "kaijuengine.com/rendering/loaders"
)

func TestKaijuMeshSerializeOmitsTriangleBVH(t *testing.T) {
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
		t.Fatal("expected explicit BVH archive generation to produce data")
	}
	data, err := km.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if !IsGLB(data) {
		t.Fatal("expected KaijuMesh.Serialize to write GLB data")
	}
	_, doc, _, err := decodeGLB(data)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Extras == nil || doc.Extras.Kaiju.Blobs != nil {
		t.Fatalf("serialized GLB should not reference triangle BVH blobs: %#v", doc.Extras)
	}
	for i := range doc.Extras.Kaiju.Meshes {
		if doc.Extras.Kaiju.Meshes[i].Blobs != nil {
			t.Fatalf("serialized GLB mesh %d should not reference triangle BVH blobs", i)
		}
	}
	db := singleAssetDatabase{key: "mesh.glb", data: data}
	if res, err := meshloaders.GLTF("mesh.glb", db); err != nil {
		t.Fatalf("serialized GLB failed standard glTF import: %v", err)
	} else if len(res.Meshes) != 1 {
		t.Fatalf("standard glTF import returned %d meshes, want 1", len(res.Meshes))
	}
	loaded, err := Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.BVH != nil {
		t.Fatal("expected serialized GLB to omit the mesh triangle BVH archive")
	}
	bvh := loaded.GenerateBVH(nil, nil, "hit")
	target, _, ok := bvh.RayIntersect(graviton.Ray{
		Origin:    matrix.Vec3{0.25, 0.25, 1},
		Direction: matrix.Vec3{0, 0, -1},
	}, 2)
	if !ok {
		t.Fatal("expected runtime bounds BVH to intersect the mesh bounds")
	}
	if target != "hit" {
		t.Fatalf("expected runtime bounds BVH data to be hydrated, got %v", target)
	}
}

func TestKaijuMeshSerializeTextureURIs(t *testing.T) {
	km := KaijuMesh{
		Name: "textured-triangle",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
			{Position: matrix.Vec3{1, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
			{Position: matrix.Vec3{0, 1, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
		},
		Indexes: []uint32{0, 1, 2},
	}
	data, err := km.SerializeWithOptions(SerializeOptions{
		TextureURIs: map[string]string{
			"baseColor":         "../texture/base.png",
			"normal":            "../texture/normal.png",
			"metallicRoughness": "../texture/mr.png",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, doc, _, err := decodeGLB(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Images) != 3 {
		t.Fatalf("expected 3 image URI entries, got %d", len(doc.Images))
	}
	got := map[string]bool{}
	for i := range doc.Images {
		got[doc.Images[i].URI] = true
	}
	for _, uri := range []string{"../texture/base.png", "../texture/normal.png", "../texture/mr.png"} {
		if !got[uri] {
			t.Fatalf("missing image URI %q in GLB images: %#v", uri, doc.Images)
		}
	}
	if loaded, err := Deserialize(data); err != nil {
		t.Fatal(err)
	} else if loaded.Textures["baseColor"] != "../texture/base.png" {
		t.Fatalf("expected baseColor texture URI to round trip, got %q", loaded.Textures["baseColor"])
	}
}

func TestKaijuMeshSetSerializeMultiMesh(t *testing.T) {
	set := KaijuMeshSet{
		Name: "multi",
		Meshes: []KaijuMesh{
			{
				Key:  "left",
				Name: "Left",
				Node: KaijuMeshNode{Name: "LeftNode", Position: matrix.Vec3{1, 0, 0}, Scale: matrix.Vec3One()},
				Verts: []rendering.Vertex{
					{Position: matrix.Vec3{0, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
					{Position: matrix.Vec3{1, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
					{Position: matrix.Vec3{0, 1, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
				},
				Indexes: []uint32{0, 1, 2},
			},
			{
				Key:  "right",
				Name: "Right",
				Node: KaijuMeshNode{Name: "RightNode", Position: matrix.Vec3{-1, 0, 0}, Scale: matrix.Vec3One()},
				Verts: []rendering.Vertex{
					{Position: matrix.Vec3{0, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
					{Position: matrix.Vec3{0, 1, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
					{Position: matrix.Vec3{-1, 0, 0}, Normal: matrix.Vec3Forward(), Color: matrix.ColorWhite()},
				},
				Indexes: []uint32{0, 1, 2},
			},
		},
	}
	data, err := set.SerializeWithOptions(SerializeOptions{
		MeshTextureURIs: map[string]map[string]string{
			"left": {"baseColor": "../texture/left.png"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !IsGLB(data) {
		t.Fatal("expected KaijuMeshSet.Serialize to write GLB data")
	}
	_, doc, _, err := decodeGLB(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Meshes) != 2 {
		t.Fatalf("GLB meshes = %d, want 2", len(doc.Meshes))
	}
	if len(doc.Nodes) < 2 {
		t.Fatalf("GLB nodes = %d, want at least 2", len(doc.Nodes))
	}
	if doc.Extras == nil || len(doc.Extras.Kaiju.Meshes) != 2 {
		t.Fatalf("extras.kaiju.meshes = %#v, want 2 entries", doc.Extras)
	}
	for i := range doc.Extras.Kaiju.Meshes {
		extra := doc.Extras.Kaiju.Meshes[i]
		if extra.Key == "" {
			t.Fatalf("extras.kaiju.meshes[%d] had no key", i)
		}
		if extra.Blobs != nil {
			t.Fatalf("extras.kaiju.meshes[%d] should not include triangle BVH blobs", i)
		}
	}
	db := singleAssetDatabase{key: "multi.glb", data: data}
	if res, err := meshloaders.GLTF("multi.glb", db); err != nil {
		t.Fatalf("serialized multi-mesh GLB failed standard glTF import: %v", err)
	} else if len(res.Meshes) != 2 {
		t.Fatalf("standard glTF import returned %d meshes, want 2", len(res.Meshes))
	}
	loadedSet, err := DeserializeSet(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(loadedSet.Meshes) != 2 {
		t.Fatalf("DeserializeSet meshes = %d, want 2", len(loadedSet.Meshes))
	}
	right, ok := loadedSet.MeshByKey("right")
	if !ok {
		t.Fatal("DeserializeSet could not find right submesh")
	}
	if right.BVH != nil {
		t.Fatal("expected right submesh GLB data to omit triangle BVH")
	}
	first, err := Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	if first.Key != "left" {
		t.Fatalf("Deserialize returned key %q, want first/default key left", first.Key)
	}
}

func TestKaijuMeshSerializeRoundTripSkinAnimation(t *testing.T) {
	rotation := matrix.QuaternionFromEuler(matrix.Vec3{0, 45, 0})
	km := KaijuMesh{
		Name: "skinned-triangle",
		Verts: []rendering.Vertex{
			{
				Position:     matrix.Vec3{0, 0, 0},
				Normal:       matrix.Vec3Forward(),
				Tangent:      matrix.Vec4{1, 0, 0, 1},
				UV0:          matrix.Vec2{0, 0},
				Color:        matrix.ColorWhite(),
				JointIds:     matrix.Vec4i{0, 1, 0, 0},
				JointWeights: matrix.Vec4{0.75, 0.25, 0, 0},
				MorphTarget:  matrix.Vec3{0, 0, 0},
			},
			{
				Position:     matrix.Vec3{1, 0, 0},
				Normal:       matrix.Vec3Forward(),
				Tangent:      matrix.Vec4{1, 0, 0, 1},
				UV0:          matrix.Vec2{1, 0},
				Color:        matrix.ColorWhite(),
				JointIds:     matrix.Vec4i{0, 1, 0, 0},
				JointWeights: matrix.Vec4{0.5, 0.5, 0, 0},
				MorphTarget:  matrix.Vec3{1, 0, 0},
			},
			{
				Position:     matrix.Vec3{0, 1, 0},
				Normal:       matrix.Vec3Forward(),
				Tangent:      matrix.Vec4{1, 0, 0, 1},
				UV0:          matrix.Vec2{0, 1},
				Color:        matrix.ColorWhite(),
				JointIds:     matrix.Vec4i{1, 0, 0, 0},
				JointWeights: matrix.Vec4{1, 0, 0, 0},
				MorphTarget:  matrix.Vec3{0, 1, 0},
			},
		},
		Indexes: []uint32{0, 1, 2},
		Joints: []KaijuMeshJoint{
			{Id: 0, Parent: -1, Skin: matrix.Mat4Identity(), Position: matrix.Vec3Zero(), Rotation: matrix.Vec3Zero(), Scale: matrix.Vec3One()},
			{Id: 1, Parent: 0, Skin: matrix.Mat4Identity(), Position: matrix.Vec3{0, 1, 0}, Rotation: matrix.Vec3Zero(), Scale: matrix.Vec3One()},
		},
		Animations: []KaijuMeshAnimation{{
			Name: "pose",
			Frames: []AnimKeyFrame{
				{
					Time: 0.5,
					Bones: []AnimBone{
						{NodeIndex: 0, PathType: AnimPathTranslation, Interpolation: AnimInterpolateLinear, Data: matrix.Vec3Zero().AsAligned16()},
						{NodeIndex: 1, PathType: AnimPathRotation, Interpolation: AnimInterpolateStep, Data: matrix.QuaternionIdentity()},
						{NodeIndex: 1, PathType: AnimPathScale, Interpolation: AnimInterpolateLinear, Data: matrix.Vec3One().AsAligned16()},
					},
				},
				{
					Time: 0,
					Bones: []AnimBone{
						{NodeIndex: 0, PathType: AnimPathTranslation, Interpolation: AnimInterpolateLinear, Data: matrix.Vec3{1, 0, 0}.AsAligned16()},
						{NodeIndex: 1, PathType: AnimPathRotation, Interpolation: AnimInterpolateStep, Data: rotation},
						{NodeIndex: 1, PathType: AnimPathScale, Interpolation: AnimInterpolateLinear, Data: matrix.Vec3{1, 2, 1}.AsAligned16()},
					},
				},
			},
		}},
	}
	data, err := km.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != km.Name {
		t.Fatalf("mesh name = %q, want %q", loaded.Name, km.Name)
	}
	if !slices.Equal(loaded.Indexes, km.Indexes) {
		t.Fatalf("indices = %#v, want %#v", loaded.Indexes, km.Indexes)
	}
	if len(loaded.Verts) != len(km.Verts) {
		t.Fatalf("verts = %d, want %d", len(loaded.Verts), len(km.Verts))
	}
	for i := range km.Verts {
		got, want := loaded.Verts[i], km.Verts[i]
		if !matrix.Vec3Approx(got.Position, want.Position) ||
			!matrix.Vec3Approx(got.Normal, want.Normal) ||
			!matrix.Vec4Approx(got.Tangent, want.Tangent) ||
			!matrix.Vec2Approx(got.UV0, want.UV0) ||
			got.Color != want.Color ||
			got.JointIds != want.JointIds ||
			!matrix.Vec4Approx(got.JointWeights, want.JointWeights) {
			t.Fatalf("vertex %d = %#v, want %#v", i, got, want)
		}
	}
	if len(loaded.Joints) != len(km.Joints) {
		t.Fatalf("joints = %d, want %d", len(loaded.Joints), len(km.Joints))
	}
	for i := range km.Joints {
		got, want := loaded.Joints[i], km.Joints[i]
		if got.Id != want.Id || got.Parent != want.Parent ||
			!matrix.Mat4Approx(got.Skin, want.Skin) ||
			!matrix.Vec3Approx(got.Position, want.Position) ||
			!matrix.Vec3Approx(got.Rotation, want.Rotation) ||
			!matrix.Vec3Approx(got.Scale, want.Scale) {
			t.Fatalf("joint %d = %#v, want %#v", i, got, want)
		}
	}
	if len(loaded.Animations) != len(km.Animations) {
		t.Fatalf("animations = %d, want %d", len(loaded.Animations), len(km.Animations))
	}
	gotAnim, wantAnim := loaded.Animations[0], km.Animations[0]
	if gotAnim.Name != wantAnim.Name || len(gotAnim.Frames) != len(wantAnim.Frames) {
		t.Fatalf("animation = %#v, want %#v", gotAnim, wantAnim)
	}
	for i := range wantAnim.Frames {
		gotFrame, wantFrame := gotAnim.Frames[i], wantAnim.Frames[i]
		if !matrix.Approx(matrix.Float(gotFrame.Time), matrix.Float(wantFrame.Time)) ||
			len(gotFrame.Bones) != len(wantFrame.Bones) {
			t.Fatalf("frame %d = %#v, want %#v", i, gotFrame, wantFrame)
		}
		for j := range wantFrame.Bones {
			gotBone, wantBone := gotFrame.Bones[j], wantFrame.Bones[j]
			if gotBone.NodeIndex != wantBone.NodeIndex ||
				gotBone.PathType != wantBone.PathType ||
				gotBone.Interpolation != wantBone.Interpolation ||
				!matrix.Vec4Approx(matrix.Vec4(gotBone.Data), matrix.Vec4(wantBone.Data)) {
				t.Fatalf("bone %d/%d = %#v, want %#v", i, j, gotBone, wantBone)
			}
		}
	}
	if loaded.BVH != nil {
		t.Fatal("expected skinned GLB round trip to omit triangle BVH")
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
	bvh := km.GenerateBVH(nil, nil, "hit")
	if bvh == nil {
		t.Fatal("expected legacy mesh data to generate a fallback BVH")
	}
	if km.BVH != nil {
		t.Fatal("expected runtime BVH generation not to populate the triangle archive")
	}
	target, _, ok := bvh.RayIntersect(graviton.Ray{
		Origin:    matrix.Vec3{0.25, 0.25, 1},
		Direction: matrix.Vec3{0, 0, -1},
	}, 2)
	if !ok {
		t.Fatal("expected runtime bounds BVH to intersect the mesh bounds")
	}
	if target != "hit" {
		t.Fatalf("expected runtime bounds BVH data to be hydrated, got %v", target)
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
