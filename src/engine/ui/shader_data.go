/******************************************************************************/
/* shader_data.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type ShaderData struct {
	rendering.ShaderDataBase
	UVs          matrix.Vec4
	FgColor      matrix.Color
	BgColor      matrix.Color
	Scissor      matrix.Vec4
	Size2D       matrix.Vec4
	BorderRadius matrix.Vec4
	BorderSize   matrix.Vec4
	BorderColor  [4]matrix.Color
	BorderLen    matrix.Vec2
	OutlineColor matrix.Color
	OutlineSize  matrix.Vec2
}

func (ShaderData) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderData{}.UVs) +
		unsafe.Sizeof(ShaderData{}.FgColor) +
		unsafe.Sizeof(ShaderData{}.BgColor) +
		unsafe.Sizeof(ShaderData{}.Scissor) +
		unsafe.Sizeof(ShaderData{}.Size2D) +
		unsafe.Sizeof(ShaderData{}.BorderRadius) +
		unsafe.Sizeof(ShaderData{}.BorderSize) +
		unsafe.Sizeof(ShaderData{}.BorderColor) +
		unsafe.Sizeof(ShaderData{}.BorderLen) +
		unsafe.Sizeof(ShaderData{}.OutlineColor) +
		unsafe.Sizeof(ShaderData{}.OutlineSize))
}

func (s *ShaderData) setUVSize(width, height matrix.Float) {
	s.UVs.SetZ(width)
	s.UVs.SetW(height)
}

func (s *ShaderData) setUVXY(x, pixelY, texSizeY matrix.Float) {
	s.UVs.SetX(x)
	s.UVs.SetY((texSizeY-pixelY)/texSizeY - s.UVs.W())
}

func (s *ShaderData) resetSize2D(ui *UI) {
	ws := ui.Entity().Transform.WorldScale()
	s.Size2D[0] = ws.X()
	s.Size2D[1] = ws.Y()
	s.Size2D[2] = ui.textureSize.X()
	s.Size2D[3] = ui.textureSize.Y()
}

func (s *ShaderData) setSize2d(ui *UI) {
	ws := ui.Entity().Transform.WorldScale()
	s.Size2D[0] = ws.X()
	s.Size2D[1] = ws.Y()
	//if matrix.Approx(s.Size2D[2], 0) {
	//	s.Size2D[2] = ui.textureSize.X()
	//}
	//if matrix.Approx(s.Size2D[3], 0) {
	//	s.Size2D[3] = ui.textureSize.Y()
	//}
}
