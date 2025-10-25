package editor_stage_manager

import (
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"slices"
)

func (m *StageManager) IsSelected(e *StageEntity) bool {
	for i := range m.selected {
		if m.selected[i] == e {
			return true
		}
	}
	return false
}

func (m *StageManager) ClearSelection() {
	for i := range m.selected {
		if sd, ok := m.selected[i].StageData.Rendering.ShaderData.(*rendering.ShaderDataStandard); ok {
			sd.ClearFlag(rendering.ShaderDataStandardFlagOutline)
		}
	}
	m.selected = klib.WipeSlice(m.selected)
}

func (m *StageManager) SelectEntity(e *StageEntity) {
	for i := range m.selected {
		if m.selected[i] == e {
			return
		}
	}
	if sd, ok := e.StageData.Rendering.ShaderData.(*rendering.ShaderDataStandard); ok {
		sd.SetFlag(rendering.ShaderDataStandardFlagOutline)
	}
	m.selected = append(m.selected, e)
}

func (m *StageManager) DeselectEntity(e *StageEntity) {
	for i := range m.selected {
		if m.selected[i] == e {
			m.selected = slices.Delete(m.selected, i, i+1)
			if sd, ok := e.StageData.Rendering.ShaderData.(*rendering.ShaderDataStandard); ok {
				sd.ClearFlag(rendering.ShaderDataStandardFlagOutline)
			}
			return
		}
	}
}

func (m *StageManager) TrySelect(ray collision.Ray) (*StageEntity, bool) {
	m.ClearSelection()
	return m.TryAppendSelect(ray)
}

func (m *StageManager) TryAppendSelect(ray collision.Ray) (*StageEntity, bool) {
	for _, e := range m.entities {
		if e.StageData.Bvh.RayIntersect(ray, matrix.FloatMax, &e.Transform) {
			m.SelectEntity(e)
			return e, true
		}
	}
	return nil, false
}

func (m *StageManager) TryToggleSelect(ray collision.Ray) (*StageEntity, bool) {
	for _, e := range m.entities {
		if e.StageData.Bvh.RayIntersect(ray, matrix.FloatMax, &e.Transform) {
			if m.IsSelected(e) {
				m.DeselectEntity(e)
			} else {
				m.SelectEntity(e)
			}
			return e, true
		}
	}
	return nil, false
}
