package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(fallback, func() rendering.DrawInstance {
		return &ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

type ShaderDataBasic struct {
	rendering.ShaderDataBase
	Color matrix.Color
}

func (t ShaderDataBasic) Size() int {
	return int(unsafe.Sizeof(ShaderDataBasic{}) - rendering.ShaderBaseDataStart)
}
