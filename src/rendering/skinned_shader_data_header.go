/******************************************************************************/
/* skinned_shader_data_header.go                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/matrix"
	"unsafe"
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
	for i := range h.bones {
		b := &h.bones[i]
		h.jointTransforms[i] = matrix.Mat4Multiply(b.Skin, b.Transform.WorldMatrix())
	}
	return skinIndex, true
}

func (h *SkinnedShaderDataHeader) isSkinNamedData(name string) bool {
	return name == "SkinnedUBO"
}
