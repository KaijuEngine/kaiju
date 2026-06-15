/******************************************************************************/
/* shader_data_grid.go                                                        */
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
		return &ShaderDataGrid{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}, "grid")
}

type ShaderDataGrid struct {
	rendering.ShaderDataBase `visible:"false"`

	Color matrix.Color
}

func (ShaderDataGrid) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataGrid{}.Color))
}
