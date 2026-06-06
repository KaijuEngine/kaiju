/******************************************************************************/
/* shader_data_particle.go                                                    */
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
		return &ShaderDataParticle{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}, "particle")
}

type ShaderDataParticle struct {
	rendering.ShaderDataBase `visible:"false"`

	Color matrix.Color
	UVs   matrix.Vec4 `default:"0,0,1,1"`
}

func (t ShaderDataParticle) Size() int {
	return int(unsafe.Sizeof(ShaderDataParticle{}) - rendering.ShaderBaseDataStart)
}
