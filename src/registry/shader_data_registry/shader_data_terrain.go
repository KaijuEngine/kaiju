/******************************************************************************/
/* shader_data_terrain.go                                                     */
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
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataTerrain{
			ShaderDataBase:        rendering.NewShaderDataBase(),
			Color:                 matrix.ColorWhite(),
			UVs:                   matrix.NewVec4(0, 0, 1, 1),
			SlopeParams:           matrix.NewVec4(0.25, 0.7, 0, 0),
			GrassTint:             matrix.NewColor(0.55, 0.72, 0.42, 1.0),
			RockTint:              matrix.NewColor(0.55, 0.52, 0.48, 1.0),
			LightDirectionAmbient: matrix.NewVec4(-0.5, -0.7, -0.5, 0.45),
			LightColorDiffuse:     matrix.NewColor(1.0, 0.95, 0.85, 1.0),
			MaterialParams:        matrix.NewVec4(1, 1, 1, 0),
			BrushColor:            matrix.NewColor(0.2, 0.75, 1.0, 1.0),
			BrushParams:           matrix.NewVec4(0.15, 0.18, 0.85, 0),
		}
	}, "terrain", "terrain_lit", "terrain_unlit", "heightScalar")
}

type ShaderDataTerrain struct {
	rendering.ShaderDataBase `visible:"false"`

	Color                 matrix.Color
	UVs                   matrix.Vec4 `default:"0,0,1,1"`
	SlopeParams           matrix.Vec4 `default:"0.25,0.7,0,0"`
	GrassTint             matrix.Color
	RockTint              matrix.Color
	LightDirectionAmbient matrix.Vec4 `default:"-0.5,-0.7,-0.5,0.45"`
	LightColorDiffuse     matrix.Color
	MaterialParams        matrix.Vec4 `default:"1,1,1,0"`
	BrushCenterRadius     matrix.Vec4 `visible:"false"`
	BrushParams           matrix.Vec4 `visible:"false" default:"0.15,0.18,0.85,0"`
	BrushColor            matrix.Color
	Flags                 StandardShaderDataFlags `visible:"false"`
}

func (t ShaderDataTerrain) Size() int {
	return int(unsafe.Sizeof(ShaderDataTerrain{}) - rendering.ShaderBaseDataStart)
}

func (t *ShaderDataTerrain) SetBrush(centerXZ matrix.Vec2, radius, ringWidth matrix.Float, color matrix.Color) {
	t.BrushCenterRadius = matrix.NewVec4(centerXZ.X(), centerXZ.Y(), radius, 1)
	t.BrushParams.SetX(matrix.Max(ringWidth, matrix.Float(0.001)))
	t.BrushColor = color
}

func (t *ShaderDataTerrain) ClearBrush() {
	t.BrushCenterRadius.SetW(0)
}
