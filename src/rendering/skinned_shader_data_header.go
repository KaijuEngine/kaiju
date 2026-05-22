/******************************************************************************/
/* skinned_shader_data_header.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"unsafe"

	"kaijuengine.com/matrix"
)

type SkinnedShaderDataHeader struct {
	bones           []BoneTransform
	boneMap         map[int32]*BoneTransform
	jointTransforms [MaxJoints]matrix.Mat4
}

type BoneTransform struct {
	Id        int32
	Transform matrix.Transform
	Skin      matrix.Mat4
}

func (h *SkinnedShaderDataHeader) HasBones() bool { return len(h.bones) > 0 }

func (h *SkinnedShaderDataHeader) CreateBones(ids []int32) {
	for i := range h.jointTransforms {
		h.jointTransforms[i].Reset()
	}
	h.bones = make([]BoneTransform, len(ids))
	h.boneMap = make(map[int32]*BoneTransform)
	for i := range ids {
		h.bones[i].Id = ids[i]
		h.boneMap[ids[i]] = &h.bones[i]
	}
}

func (h *SkinnedShaderDataHeader) BoneByIndex(index int) *BoneTransform {
	return &h.bones[index]
}

func (h *SkinnedShaderDataHeader) FindBone(id int32) *BoneTransform {
	if id < 0 {
		return nil
	}
	if b, ok := h.boneMap[id]; ok {
		return b
	}
	return nil
}

func (h *SkinnedShaderDataHeader) SkinNamedDataInstanceSize() int {
	return int(unsafe.Sizeof(h.jointTransforms))
}

func (h *SkinnedShaderDataHeader) SkinNamedDataPointer() unsafe.Pointer {
	return unsafe.Pointer(&h.jointTransforms)
}

func (h *SkinnedShaderDataHeader) SkinUpdateNamedData() bool {
	for i := range h.bones {
		b := &h.bones[i]
		h.jointTransforms[i] = matrix.Mat4Multiply(b.Skin, b.Transform.WorldMatrix())
	}
	return true
}
