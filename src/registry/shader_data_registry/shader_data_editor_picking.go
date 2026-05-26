/******************************************************************************/
/* shader_data_editor_picking.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"unsafe"

	"kaijuengine.com/rendering"
)

func init() {
	create := func() rendering.DrawInstance {
		return &ShaderDataEditorPicking{
			ShaderDataBase: rendering.NewShaderDataBase(),
		}
	}
	register(create, "editor_pick")
	register(create, "editor_picking")
}

type ShaderDataEditorPicking struct {
	rendering.ShaderDataBase `visible:"false"`

	PickID uint32 `visible:"false"`
}

func (t ShaderDataEditorPicking) Size() int {
	return int(unsafe.Sizeof(ShaderDataEditorPicking{}) - rendering.ShaderBaseDataStart)
}
