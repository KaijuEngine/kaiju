/******************************************************************************/
/* shader_data_basic_unlit.go                                                 */
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
		return &ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}, "unlit", "unlit_transparent")
}

type ShaderDataUnlit struct {
	rendering.ShaderDataBase `visible:"false"`

	Color matrix.Color
	UVs   matrix.Vec4             `default:"0,0,1,1"`
	Flags StandardShaderDataFlags `visible:"false"`
}

func (t ShaderDataUnlit) Size() int {
	return int(unsafe.Sizeof(ShaderDataUnlit{}) - rendering.ShaderBaseDataStart)
}
