/******************************************************************************/
/* shader_data_ed_transform_wire.go                                           */
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
		return &ShaderDataEdFrustumWire{
			ShaderDataBase:    rendering.NewShaderDataBase(),
			Color:             matrix.ColorWhite(),
			FrustumProjection: matrix.Mat4Identity(),
		}
	}, "ed_frustum_wire")
}

type ShaderDataEdFrustumWire struct {
	rendering.ShaderDataBase `visible:"false"`

	Color             matrix.Color
	FrustumProjection matrix.Mat4
}

func (ShaderDataEdFrustumWire) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataEdFrustumWire{}.Color) +
		unsafe.Sizeof(ShaderDataEdFrustumWire{}.FrustumProjection))
}
