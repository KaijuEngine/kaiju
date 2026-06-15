/******************************************************************************/
/* shader_data_standard_skinned.go                                            */
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
		return &ShaderDataStandardSkinned{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorWhite(),
		}
	}, fallback+"_skinned")
}

type ShaderDataStandardSkinned struct {
	rendering.SkinnedShaderDataHeader `visible:"false"`
	rendering.ShaderDataBase          `visible:"false"`

	Color matrix.Color
	Flags StandardShaderDataFlags `visible:"false"`
}

func (ShaderDataStandardSkinned) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataStandardSkinned{}.Color) +
		unsafe.Sizeof(ShaderDataStandardSkinned{}.Flags))
}

func (t *ShaderDataStandardSkinned) SkinningHeader() *rendering.SkinnedShaderDataHeader {
	return &t.SkinnedShaderDataHeader
}

func (t *ShaderDataStandardSkinned) InstanceBoundDataSize() int {
	return t.SkinNamedDataInstanceSize()
}

func (t *ShaderDataStandardSkinned) BoundDataPointer() unsafe.Pointer {
	return t.SkinNamedDataPointer()
}

func (t *ShaderDataStandardSkinned) UpdateBoundData() bool {
	return t.SkinUpdateNamedData()
}
