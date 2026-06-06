/******************************************************************************/
/* skinned_shader_data_header_test.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"unsafe"

	"kaijuengine.com/matrix"
)

func TestSkinnedShaderDataHeaderCreateBones(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	header.CreateBones([]int32{10, 20})
	if !header.HasBones() || len(header.bones) != 2 || len(header.boneMap) != 2 {
		t.Fatalf("bones were not created: %+v", header)
	}
	if header.bones[0].Id != 10 || header.bones[1].Id != 20 {
		t.Fatalf("bone IDs = %+v", header.bones)
	}
	for i := range header.jointTransforms {
		if header.jointTransforms[i] != matrix.Mat4Identity() {
			t.Fatalf("joint transform %d was not reset", i)
		}
	}
}

func TestSkinnedShaderDataHeaderBoneLookup(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	header.CreateBones([]int32{10, 20})
	if header.BoneByIndex(1).Id != 20 {
		t.Fatalf("BoneByIndex returned wrong bone")
	}
	if header.FindBone(10) != &header.bones[0] {
		t.Fatalf("FindBone did not return mapped bone")
	}
	if header.FindBone(99) != nil || header.FindBone(-1) != nil {
		t.Fatalf("missing and negative bone lookups should be nil")
	}
}

func TestSkinnedShaderDataHeaderNamedData(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	if header.HasBones() {
		t.Fatalf("empty header should not have bones")
	}
	if header.SkinNamedDataInstanceSize() != int(unsafe.Sizeof(header.jointTransforms)) {
		t.Fatalf("SkinNamedDataInstanceSize = %d", header.SkinNamedDataInstanceSize())
	}
	if header.SkinNamedDataPointer() != unsafe.Pointer(&header.jointTransforms) {
		t.Fatalf("SkinNamedDataPointer should point at jointTransforms")
	}
}

func TestSkinnedShaderDataHeaderSkinUpdateNamedData(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	header.CreateBones([]int32{10})
	bone := header.BoneByIndex(0)
	bone.Transform.SetupRawTransform()
	bone.Transform.SetPosition(matrix.Vec3{1, 2, 3})
	bone.Skin = matrix.Mat4Identity()
	bone.Skin.Translate(matrix.Vec3{4, 0, 0})
	if !header.SkinUpdateNamedData() {
		t.Fatalf("SkinUpdateNamedData should return true")
	}
	want := matrix.Mat4Multiply(bone.Skin, bone.Transform.WorldMatrix())
	if header.jointTransforms[0] != want {
		t.Fatalf("joint transform = %v, want %v", header.jointTransforms[0], want)
	}
}
