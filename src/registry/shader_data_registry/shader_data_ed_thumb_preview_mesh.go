/******************************************************************************/
/* shader_data_ed_thumb_preview_mesh.go                                       */
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
