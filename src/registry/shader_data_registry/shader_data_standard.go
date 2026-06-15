/******************************************************************************/
/* shader_data_standard.go                                                    */
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
		return &ShaderDataStandard{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}, fallback)
}

type ShaderDataStandard struct {
	rendering.ShaderDataBase `visible:"false"`

	Color matrix.Color
	Flags StandardShaderDataFlags `visible:"false"`
}

func (ShaderDataStandard) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataStandard{}.Color) +
		unsafe.Sizeof(ShaderDataStandard{}.Flags))
}
