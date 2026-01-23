package shader_data_registry

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"unsafe"
)

func init() {
	register(func() rendering.DrawInstance {
		return &ShaderDataEdThumbPreviewMesh{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorSlateGrey(),
		}
	}, "ed_thumb_preview_mesh")
}

type ShaderDataEdThumbPreviewMesh struct {
	rendering.ShaderDataBase `visible:"false"`

	Color matrix.Color
	Flags StandardShaderDataFlags `visible:"false"`
}

func (ShaderDataEdThumbPreviewMesh) Size() int {
	return int(unsafe.Sizeof(ShaderDataEdThumbPreviewMesh{}) - rendering.ShaderBaseDataStart)
}
