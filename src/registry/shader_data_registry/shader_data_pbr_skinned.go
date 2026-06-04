/******************************************************************************/
/* shader_data_pbr_skinned.go                                                 */
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
		return &ShaderDataPbrSkinned{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			MeRoEmAo:       matrix.NewVec4(1, 1, 0, 1),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	}, "pbr_skinned")
}

type ShaderDataPbrSkinned struct {
	rendering.SkinnedShaderDataHeader `visible:"false"`
	rendering.ShaderDataBase          `visible:"false"`

	VertColors matrix.Color
	MeRoEmAo   matrix.Vec4
	Flags      StandardShaderDataFlags `visible:"false"`
	LightIds   [4]int32                `visible:"false"`
}

func (t *ShaderDataPbrSkinned) SkinningHeader() *rendering.SkinnedShaderDataHeader {
	return &t.SkinnedShaderDataHeader
}

func (t ShaderDataPbrSkinned) Size() int {
	return int(unsafe.Sizeof(ShaderDataPbrSkinned{}) - rendering.ShaderBaseDataStart)
}

func (s *ShaderDataPbrSkinned) SelectLights(lights rendering.LightsForRender) {
	selectPBRLights(&s.ShaderDataBase, &s.LightIds, lights)
}

func (t *ShaderDataPbrSkinned) InstanceBoundDataSize() int {
	return t.SkinNamedDataInstanceSize()
}

func (t *ShaderDataPbrSkinned) BoundDataPointer() unsafe.Pointer {
	return t.SkinNamedDataPointer()
}

func (t *ShaderDataPbrSkinned) UpdateBoundData() bool {
	return t.SkinUpdateNamedData()
}
