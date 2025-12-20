/******************************************************************************/
/* shader_data_pbr.go                                                         */
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
		return &ShaderDataPBR{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	}, "pbr")
}

type ShaderDataPBR struct {
	rendering.ShaderDataBase
	VertColors matrix.Color
	Metallic   float32
	Roughness  float32
	Emissive   float32
	LightIds   [4]int32 `visible:"false"`
}

func (t ShaderDataPBR) Size() int {
	return int(unsafe.Sizeof(ShaderDataPBR{}) - rendering.ShaderBaseDataStart)
}

func (s *ShaderDataPBR) SelectLights(lights rendering.LightsForRender) {
	shouldUpdate := lights.HasChanges
	t := s.Transform()
	shouldUpdate = shouldUpdate || (t != nil && t.IsDirty())
	if !shouldUpdate {
		return
	}
	// TODO:  This is for testing, should select closest
	for i := range s.LightIds {
		s.LightIds[i] = -1
	}
	for i := range lights.Lights {
		if lights.Lights[i].IsValid() {
			s.LightIds[i] = int32(i)
		} else {
			break
		}
	}
}
