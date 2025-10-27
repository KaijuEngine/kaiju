package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(fallback+"_lit", func() rendering.DrawInstance {
		return &ShaderDataBasicLit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

type ShaderDataBasicLit struct {
	rendering.ShaderDataBase
	Color  matrix.Color
	Light0 float32
	Light1 float32
	Light2 float32
	Light3 float32
}

func (t ShaderDataBasicLit) Size() int {
	return int(unsafe.Sizeof(ShaderDataBasicLit{}) - rendering.ShaderBaseDataStart)
}
