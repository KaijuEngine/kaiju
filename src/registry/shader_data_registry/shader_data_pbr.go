/******************************************************************************/
/* shader_data_pbr.go                                                         */
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
		return &ShaderDataPBR{
			ShaderDataBase: rendering.NewShaderDataBase(),
			VertColors:     matrix.ColorWhite(),
			MeRoEmAo:       matrix.NewVec4(1, 1, 0, 1),
			LightIds:       [...]int32{-1, -1, -1, -1},
		}
	}, "pbr")
}

type ShaderDataPBR struct {
	rendering.ShaderDataBase `visible:"false"`

	VertColors matrix.Color
	MeRoEmAo   matrix.Vec4
	Flags      StandardShaderDataFlags `visible:"false"`
	LightIds   [4]int32                `visible:"false"`
}

func (t ShaderDataPBR) Size() int {
	return int(unsafe.Sizeof(ShaderDataPBR{}) - rendering.ShaderBaseDataStart)
}

func (s *ShaderDataPBR) SelectLights(lights rendering.LightsForRender) {
	selectPBRLights(&s.ShaderDataBase, &s.LightIds, lights)
}

func selectPBRLights(base *rendering.ShaderDataBase, ids *[4]int32, lights rendering.LightsForRender) {
	shouldUpdate := lights.HasChanges
	t := base.Transform()
	shouldUpdate = shouldUpdate || (t != nil && t.IsDirty())
	if !shouldUpdate {
		return
	}
	for i := range ids {
		ids[i] = -1
	}
	slot := 0
	for i := range lights.Lights {
		if slot >= len(ids) {
			break
		}
		if lights.Lights[i].IsValid() {
			ids[slot] = int32(i)
			slot++
		}
	}
}
