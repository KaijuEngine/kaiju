/******************************************************************************/
/* editor_stage_manager_interact.go                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor_stage_manager

import (
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"slices"
)

func (m *StageManager) HasSelection() bool { return len(m.selected) > 0 }

func (m *StageManager) IsSelected(e *StageEntity) bool {
	defer tracing.NewRegion("StageManager.IsSelected").End()
	for i := range m.selected {
		if m.selected[i] == e {
			return true
		}
	}
	return false
}

func (m *StageManager) ClearSelection() {
	defer tracing.NewRegion("StageManager.ClearSelection").End()
	for i := range m.selected {
		if sd, ok := m.selected[i].StageData.ShaderData.(*shader_data_registry.ShaderDataStandard); ok {
			sd.ClearFlag(shader_data_registry.ShaderDataStandardFlagOutline)
		}
	}
	m.selected = klib.WipeSlice(m.selected)
}

func (m *StageManager) SelectEntity(e *StageEntity) {
	defer tracing.NewRegion("StageManager.SelectEntity").End()
	for i := range m.selected {
		if m.selected[i] == e {
			return
		}
	}
	if sd, ok := e.StageData.ShaderData.(*shader_data_registry.ShaderDataStandard); ok {
		sd.SetFlag(shader_data_registry.ShaderDataStandardFlagOutline)
	}
	m.selected = append(m.selected, e)
}

func (m *StageManager) DeselectEntity(e *StageEntity) {
	defer tracing.NewRegion("StageManager.DeselectEntity").End()
	for i := range m.selected {
		if m.selected[i] == e {
			m.selected = slices.Delete(m.selected, i, i+1)
			if sd, ok := e.StageData.ShaderData.(*shader_data_registry.ShaderDataStandard); ok {
				sd.ClearFlag(shader_data_registry.ShaderDataStandardFlagOutline)
			}
			return
		}
	}
}

func (m *StageManager) TryHitEntity(ray collision.Ray) (*StageEntity, bool) {
	for _, e := range m.entities {
		if e.StageData.Bvh.RayIntersect(ray, matrix.FloatMax, &e.Transform) {
			return e, true
		}
	}
	return nil, false
}

func (m *StageManager) TrySelect(ray collision.Ray) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.TrySelect").End()
	m.ClearSelection()
	return m.TryAppendSelect(ray)
}

func (m *StageManager) TryAppendSelect(ray collision.Ray) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.TryAppendSelect").End()
	if e, ok := m.TryHitEntity(ray); ok {
		m.SelectEntity(e)
		return e, true
	}
	return nil, false
}

func (m *StageManager) TryToggleSelect(ray collision.Ray) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.TryToggleSelect").End()
	if e, ok := m.TryHitEntity(ray); ok {
		if m.IsSelected(e) {
			m.DeselectEntity(e)
		} else {
			m.SelectEntity(e)
		}
		return e, true
	}
	return nil, false
}

func (m *StageManager) SelectionCenter() matrix.Vec3 {
	defer tracing.NewRegion("StageManager.SelectionCenter").End()
	center := matrix.Vec3Zero()
	for _, e := range m.selected {
		b := e.StageData.Bvh.Bounds(&e.Transform)
		center.AddAssign(b.Center)
	}
	return center
}

func (m *StageManager) SelectionPivotCenter() matrix.Vec3 {
	defer tracing.NewRegion("StageManager.SelectionPivotCenter").End()
	center := matrix.Vec3Zero()
	for _, e := range m.selected {
		center.AddAssign(e.Transform.WorldPosition())
	}
	return center
}

func (m *StageManager) SelectionBounds() collision.AABB {
	defer tracing.NewRegion("StageManager.SelectionBounds").End()
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
