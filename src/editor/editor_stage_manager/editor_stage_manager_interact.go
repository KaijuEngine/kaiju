package editor_stage_manager

import (
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
)

func (m *StageManager) ClearSelection() {
	for i := range m.selected {
		if sd, ok := m.selected[i].StageData.Rendering.ShaderData.(*rendering.ShaderDataStandard); ok {
			sd.ClearFlag(rendering.ShaderDataStandardFlagOutline)
		}
	}
	m.selected = klib.WipeSlice(m.selected)
}

func (m *StageManager) TrySelect(ray collision.Ray) (*StageEntity, bool) {
	m.ClearSelection()
	for _, e := range m.entities {
		if e.StageData.Bvh.RayIntersect(ray, matrix.FloatMax, &e.Transform) {
			if sd, ok := e.StageData.Rendering.ShaderData.(*rendering.ShaderDataStandard); ok {
				sd.SetFlag(rendering.ShaderDataStandardFlagOutline)
			}
			m.selected = append(m.selected, e)
			return e, true
		}
	}
	return nil, false
}
