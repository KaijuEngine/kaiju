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
		}
	}, "ed_thumb_preview_mesh")
}

type ShaderDataEdThumbPreviewMesh struct {
	rendering.ShaderDataBase `visible:"false"`

	View       matrix.Mat4 `visible:"false"`
	Projection matrix.Mat4 `visible:"false"`
}

func (s *ShaderDataEdThumbPreviewMesh) SetCamera(view, projection matrix.Mat4) {
	s.View = view
	s.Projection = projection
}

func (ShaderDataEdThumbPreviewMesh) Size() int {
	return int(unsafe.Sizeof(ShaderDataEdThumbPreviewMesh{}) - rendering.ShaderBaseDataStart)
}
