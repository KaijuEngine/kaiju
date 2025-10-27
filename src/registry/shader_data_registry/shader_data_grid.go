package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register("grid", func() rendering.DrawInstance {
		return &ShaderDataGrid{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

type ShaderDataGrid struct {
	rendering.ShaderDataBase
	Color matrix.Color
}

func (t ShaderDataGrid) Size() int {
	return int(unsafe.Sizeof(ShaderDataGrid{}) - rendering.ShaderBaseDataStart)
}
