package editor_stage_manager

import (
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"slices"
)

func (m *StageManager) HasSelection() bool { return len(m.selected) > 0 }

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

func (m *StageManager) SelectionBounds() collision.AABB {
	low := matrix.Vec3Inf(1)
	high := matrix.Vec3Inf(-1)
	center := matrix.Vec3Zero()
	for _, e := range m.selected {
		p := e.Transform.Position()
		b := e.StageData.Bvh.Bounds(&e.Transform)
		center.AddAssign(b.Center)
		ex := matrix.Vec3Max(matrix.Vec3Zero(), b.Extent)
		low = matrix.Vec3Min(low, p.Subtract(ex))
		high = matrix.Vec3Max(high, p.Add(ex))
	}
	center.ShrinkAssign(float32(len(m.selected)))
	return collision.AABB{
		Center: center,
		Extent: high.Subtract(low).Scale(0.5),
	}
}
