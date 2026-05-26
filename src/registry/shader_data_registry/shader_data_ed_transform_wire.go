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
		return &ShaderDataEdTransformWire{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}, "ed_transform_wire")
}

type ShaderDataEdTransformWire struct {
	rendering.ShaderDataBase `visible:"false"`

	Color matrix.Color
}

func (t ShaderDataEdTransformWire) Size() int {
	return int(unsafe.Sizeof(ShaderDataEdTransformWire{}) - rendering.ShaderBaseDataStart)
}
