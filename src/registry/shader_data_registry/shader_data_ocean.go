/******************************************************************************/
/* shader_data_ocean.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataOcean{
			ShaderDataBase: rendering.NewShaderDataBase(),
			ShallowColor:   matrix.NewColor(0.055, 0.42, 0.53, 1),
			DeepColor:      matrix.NewColor(0.008, 0.09, 0.18, 1),
			WaveParams:     matrix.NewVec4(0.08, 0.52, 0.58, 0.45),
			LightIds:       [...]int32{-1, -1, -1, -1},
			BrushParams:    matrix.NewVec4(0.15, 0.18, 0.85, 0),
			BrushColor:     matrix.NewColor(0.2, 0.75, 1.0, 1.0),
		}
	}, "ocean")
}

// ShaderDataOcean controls a procedural, nonmetallic water surface.
// WaveParams are amplitude, speed, spatial frequency, and roughness.
type ShaderDataOcean struct {
	rendering.ShaderDataBase `visible:"false"`

	ShallowColor      matrix.Color
	DeepColor         matrix.Color
	WaveParams        matrix.Vec4
	Flags             StandardShaderDataFlags `visible:"false"`
	LightIds          [4]int32                `visible:"false"`
	BrushCenterRadius matrix.Vec4             `visible:"false"`
	BrushParams       matrix.Vec4             `visible:"false"`
	BrushColor        matrix.Color
}

func (ShaderDataOcean) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataOcean{}.ShallowColor) +
		unsafe.Sizeof(ShaderDataOcean{}.DeepColor) +
		unsafe.Sizeof(ShaderDataOcean{}.WaveParams) +
		unsafe.Sizeof(ShaderDataOcean{}.Flags) +
		unsafe.Sizeof(ShaderDataOcean{}.LightIds) +
		unsafe.Sizeof(ShaderDataOcean{}.BrushCenterRadius) +
		unsafe.Sizeof(ShaderDataOcean{}.BrushParams) +
		unsafe.Sizeof(ShaderDataOcean{}.BrushColor))
}

func (s *ShaderDataOcean) SelectLights(lights rendering.LightsForRender) {
	selectPBRLights(&s.ShaderDataBase, &s.LightIds, lights)
}

func (s *ShaderDataOcean) SetBrush(centerXZ matrix.Vec2, radius, ringWidth matrix.Float, color matrix.Color) {
	s.BrushCenterRadius = matrix.NewVec4(centerXZ.X(), centerXZ.Y(), radius, 1)
	s.BrushParams.SetX(max(ringWidth, matrix.Float(0.001)))
	s.BrushColor = color
}

func (s *ShaderDataOcean) ClearBrush() {
	s.BrushCenterRadius.SetW(0)
}
