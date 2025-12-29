package rendering

import (
	"kaiju/matrix"
	"unsafe"
)

type SkinnedShaderDataHeader struct {
	bones           []BoneTransform
	jointTransforms [MaxJoints]matrix.Mat4
}

type BoneTransform struct {
	Id        int32
	Transform matrix.Transform
	Skin      matrix.Mat4
}

func (h *SkinnedShaderDataHeader) CreateBones(count int) {
	h.bones = make([]BoneTransform, count)
}

func (h *SkinnedShaderDataHeader) Bone(index int) *BoneTransform {
	return &h.bones[index]
}

func (h *SkinnedShaderDataHeader) FindBone(id int32) *BoneTransform {
	if id < 0 {
		return nil
	}
	for i := range h.bones {
		if h.bones[i].Id == id {
			return &h.bones[i]
		}
	}
	return nil
}

func (h *SkinnedShaderDataHeader) SkinNamedDataInstanceSize(name string) int {
	if !h.isSkinNamedData(name) {
		return 0
	}
	return int(unsafe.Sizeof(h.jointTransforms))
}

func (h *SkinnedShaderDataHeader) SkinNamedDataPointer(name string) unsafe.Pointer {
	if !h.isSkinNamedData(name) {
		return nil
	}
	return unsafe.Pointer(&h.jointTransforms)
}

func (h *SkinnedShaderDataHeader) SkinUpdateNamedData(index, capacity int, name string) (int32, bool) {
	if !h.isSkinNamedData(name) {
		return 0, false
	}
	cap := capacity / MaxJoints / int(unsafe.Sizeof(matrix.Mat4{}))
	if index > cap {
		return int32(index % cap), false
	}
	skinIndex := int32(index)
	if len(h.bones) > 0 {
		for i := range h.bones {
			b := &h.bones[i]
			m := matrix.Mat4Multiply(b.Skin, b.Transform.Matrix())
			parent := b.Transform.Parent()
			for parent != nil {
				m.MultiplyAssign(parent.Matrix())
				parent = parent.Parent()
			}
			h.jointTransforms[i] = m
		}
	}
	return skinIndex, true
}

func (h *SkinnedShaderDataHeader) isSkinNamedData(name string) bool {
	return name == "SkinnedUBO"
}
