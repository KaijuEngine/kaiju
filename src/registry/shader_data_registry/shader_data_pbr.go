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
	shouldUpdate := lights.HasChanges
	t := s.Transform()
	shouldUpdate = shouldUpdate || (t != nil && t.IsDirty())
	if !shouldUpdate {
		return
	}
	// TODO:  This is for testing, should select closest
	for i := range s.LightIds {
		s.LightIds[i] = -1
	}
	for i := range lights.Lights {
		if lights.Lights[i].IsValid() {
			s.LightIds[i] = int32(i)
		} else {
			break
		}
	}
}
