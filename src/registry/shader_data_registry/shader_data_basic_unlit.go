package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(fallback+"_unlit", func() rendering.DrawInstance {
		return &ShaderDataUnlit{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

type ShaderDataUnlit struct {
	rendering.ShaderDataBase
	Color matrix.Color
	UVs   matrix.Vec4
}

func (t ShaderDataUnlit) Size() int {
	return int(unsafe.Sizeof(ShaderDataUnlit{}) - rendering.ShaderBaseDataStart)
}
