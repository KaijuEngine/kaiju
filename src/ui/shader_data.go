package ui

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
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
}

func (s ShaderData) Size() int {
	const size = int(unsafe.Sizeof(ShaderData{}) - rendering.ShaderBaseDataStart)
	return size
}

func (s *ShaderData) setSize2d(ui UI, texWidth, texHeight float32) {
	// TODO:  This is skipped for text
	ws := ui.Entity().Transform.WorldScale()
	s.Size2D[0] = ws.X()
	s.Size2D[1] = ws.Y()
	s.Size2D[2] = texWidth
	s.Size2D[3] = texHeight
}
