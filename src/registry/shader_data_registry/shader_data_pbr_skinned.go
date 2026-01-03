/******************************************************************************/
/* shader_data_pbr_skinned.go                                                 */
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

package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataPbrSkinned{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	}, "pbr_skinned")
}

type ShaderDataPbrSkinned struct {
	rendering.SkinnedShaderDataHeader `visible:"false"`
	rendering.ShaderDataBase          `visible:"false"`

	VertColors matrix.Color
	Metallic   float32
	Roughness  float32
	Emissive   float32
	LightIds   [4]int32                `visible:"false"`
	SkinIndex  int32                   `visible:"false"`
	Flags      StandardShaderDataFlags `visible:"false"`
}

func (t *ShaderDataPbrSkinned) SkinningHeader() *rendering.SkinnedShaderDataHeader {
	return &t.SkinnedShaderDataHeader
}

func (t ShaderDataPbrSkinned) Size() int {
	const size = int(unsafe.Sizeof(ShaderDataPbrSkinned{}) - rendering.ShaderBaseDataStart)
	return size
}

func (t *ShaderDataPbrSkinned) NamedDataInstanceSize(name string) int {
	return t.SkinNamedDataInstanceSize(name)
}

func (t *ShaderDataPbrSkinned) NamedDataPointer(name string) unsafe.Pointer {
	return t.SkinNamedDataPointer(name)
}

func (t *ShaderDataPbrSkinned) UpdateNamedData(index, _ int, name string) bool {
	if t.SkinUpdateNamedData(name) {
		t.SkinIndex = int32(index)
		return true
	}
	return false
}

func (s *ShaderDataPbrSkinned) TestFlag(flag StandardShaderDataFlags) bool {
	return (s.Flags & flag) != 0
}

func (s *ShaderDataPbrSkinned) SetFlag(flag StandardShaderDataFlags) {
	s.Flags |= flag
	s.updateFlagEnableStatus()
}

func (s *ShaderDataPbrSkinned) ClearFlag(flag StandardShaderDataFlags) {
	s.Flags &^= flag
	s.updateFlagEnableStatus()
}

func (s *ShaderDataPbrSkinned) updateFlagEnableStatus() {
	if s.Flags|ShaderDataStandardFlagEnable == ShaderDataStandardFlagEnable {
		s.Flags = 0
	} else {
		s.Flags |= ShaderDataStandardFlagEnable
	}
}
