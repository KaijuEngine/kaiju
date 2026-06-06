/******************************************************************************/
/* shader_draw.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"sort"
	"unsafe"

	"kaijuengine.com/klib"
)

type ShaderDraw struct {
	material         *Material
	instanceGroups   []DrawInstanceGroup
	pushConstantData unsafe.Pointer
}

func NewShaderDraw(material *Material) ShaderDraw {
	return ShaderDraw{
		material:       material,
		instanceGroups: make([]DrawInstanceGroup, 0),
	}
}

func (s *ShaderDraw) AddInstanceGroup(group DrawInstanceGroup) {
	s.instanceGroups = append(s.instanceGroups, group)
	sort.Slice(s.instanceGroups, func(i, j int) bool {
		return s.instanceGroups[i].sort < s.instanceGroups[j].sort
	})
}

func (s *ShaderDraw) Filter(filter func(*DrawInstanceGroup) bool) []*DrawInstanceGroup {
	selected := make([]*DrawInstanceGroup, 0, len(s.instanceGroups))
	for i := range s.instanceGroups {
		if filter(&s.instanceGroups[i]) {
			selected = append(selected, &s.instanceGroups[i])
		}
	}
	return selected
}

func (s *ShaderDraw) Destroy(device *GPUDevice) {
	for i := range s.instanceGroups {
		s.instanceGroups[i].Destroy(device)
	}
	s.material = nil
	s.instanceGroups = klib.WipeSlice(s.instanceGroups)
}

func (s *ShaderDraw) Clear() {
	for i := range s.instanceGroups {
		s.instanceGroups[i].Clear()
	}
}
