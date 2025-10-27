package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register("pbr", func() rendering.DrawInstance {
		return &ShaderDataPBR{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
		}
	})
}

type ShaderDataPBR struct {
	rendering.ShaderDataBase
	VertColors matrix.Color
	Metallic   float32
	Roughness  float32
	Emissive   float32
	Light0     float32
	Light1     float32
	Light2     float32
	Light3     float32
}

func (t ShaderDataPBR) Size() int {
	return int(unsafe.Sizeof(ShaderDataPBR{}) - rendering.ShaderBaseDataStart)
}
