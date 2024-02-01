package sprite

import (
	"kaiju/matrix"
	"kaiju/rendering"
)

type ShaderData struct {
	rendering.ShaderDataBase
	UVs       matrix.Vec4
	FgColor   matrix.Color
	BgColor   matrix.Color
	Scissor   matrix.Vec4
	Size2D    matrix.Vec4
	BorderLen matrix.Vec2
}
