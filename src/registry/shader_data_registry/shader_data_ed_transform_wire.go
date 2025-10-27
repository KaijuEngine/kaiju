package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register("ed_transform_wire", func() rendering.DrawInstance {
		return &ShaderDataEdTransformWire{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	})
}

type ShaderDataEdTransformWire struct {
	rendering.ShaderDataBase
	Color matrix.Color
}

func (t ShaderDataEdTransformWire) Size() int {
	return int(unsafe.Sizeof(ShaderDataGrid{}) - rendering.ShaderBaseDataStart)
}
