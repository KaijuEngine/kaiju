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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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

func (m *StageManager) IsSelectedById(id string) bool {
	defer tracing.NewRegion("StageManager.IsSelected").End()
	for i := range m.selected {
		if m.selected[i].StageData.Description.Id == id {
			return true
		}
	}
	return false
}

func (m *StageManager) ClearSelection() {
	defer tracing.NewRegion("StageManager.ClearSelection").End()
	if len(m.selected) == 0 {
		return
	}
	cpy := slices.Clone(m.selected)
	m.selected = klib.WipeSlice(m.selected)
	m.history.Add(&selectHistory{
		manager: m,
		from:    cpy,
	})
	for i := range cpy {
		m.clearShaderDataFlag(cpy[i])
		m.OnEntityDeselected.Execute(cpy[i])
	}
}

func (m *StageManager) SelectEntity(e *StageEntity) {
	defer tracing.NewRegion("StageManager.SelectEntity").End()
	for i := range m.selected {
		if m.selected[i] == e {
			return
		}
	}
	history := &selectHistory{
		manager: m,
		from:    slices.Clone(m.selected),
	}
	m.selected = append(m.selected, e)
	history.to = slices.Clone(m.selected)
	m.history.Add(history)
	m.setShaderDataFlag(e)
	m.OnEntitySelected.Execute(e)
}

func (m *StageManager) SelectEntityById(id string) {
	m.ClearSelection()
	m.SelectAppendEntityById(id)
}

func (m *StageManager) SelectAppendEntityById(id string) {
	if e, ok := m.EntityById(id); ok {
		m.SelectEntity(e)
	}
}

func (m *StageManager) SelectToggleEntityById(id string) {
	if e, ok := m.EntityById(id); ok {
		if m.IsSelected(e) {
			m.DeselectEntity(e)
		} else {
			m.SelectEntity(e)
		}
	}
}

func (m *StageManager) DeselectEntity(e *StageEntity) {
	defer tracing.NewRegion("StageManager.DeselectEntity").End()
	for i := range m.selected {
		if m.selected[i] == e {
			history := &selectHistory{
				manager: m,
				from:    slices.Clone(m.selected),
			}
			m.selected = slices.Delete(m.selected, i, i+1)
			history.to = slices.Clone(m.selected)
			m.clearShaderDataFlag(e)
			m.OnEntityDeselected.Execute(e)
			return
		}
	}
}

func (m *StageManager) TryHitEntity(ray collision.Ray) (*StageEntity, matrix.Vec3, bool) {
	if target, pt, ok := m.worldBVH.RayIntersect(ray, 1000); ok {
		return target.(*StageEntity), pt, ok
	}
	return nil, matrix.Vec3{}, false
}

func (m *StageManager) TrySelect(ray collision.Ray) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.TrySelect").End()
	m.ClearSelection()
	return m.TryAppendSelect(ray)
}

func (m *StageManager) TryBoxSelect(screenBox matrix.Vec4) {
	defer tracing.NewRegion("StageManager.TryBoxSelect").End()
	m.history.BeginTransaction()
	defer m.history.CommitTransaction()
	m.ClearSelection()
	cam := m.host.PrimaryCamera()
	f := cam.Frustum()
	v, p := cam.View(), cam.Projection()
	viewport := cam.Viewport()
	for _, e := range m.entities {
		if e.StageData.Bvh == nil || e.isDeleted {
			continue
		}
		b := e.StageData.Bvh.Bounds()
		if !b.IntersectsFrustum(f) {
			continue
		}
		ss, ok := matrix.Mat4ToScreenSpace(e.Transform.Position(), v, p, viewport)
		if !ok {
			continue
		}
		if screenBox.AreaContains(ss.X(), ss.Y()) {
			m.SelectEntity(e)
		}
	}
}

func (m *StageManager) TryAppendSelect(ray collision.Ray) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.TryAppendSelect").End()
	if e, _, ok := m.TryHitEntity(ray); ok {
		m.SelectEntity(e)
		return e, true
	}
	return nil, false
}

func (m *StageManager) TryToggleSelect(ray collision.Ray) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.TryToggleSelect").End()
	if e, _, ok := m.TryHitEntity(ray); ok {
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
		b := e.StageData.Bvh.Bounds()
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
		var b collision.AABB
		if e.StageData.Bvh != nil {
			b = e.StageData.Bvh.Bounds()
			b.Center.AddAssign(p)
		} else {
			b = collision.AABBFromTransform(&e.Transform)
		}
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

func (m *StageManager) setShaderDataFlag(root *StageEntity) {
	var procChildren func(e *StageEntity)
	procChildren = func(e *StageEntity) {
		switch sd := e.StageData.ShaderData.(type) {
		case *shader_data_registry.ShaderDataStandard:
			sd.SetFlag(shader_data_registry.ShaderDataStandardFlagOutline)
		case *shader_data_registry.ShaderDataStandardSkinned:
			sd.SetFlag(shader_data_registry.ShaderDataStandardFlagOutline)
		case *shader_data_registry.ShaderDataPBR:
			sd.SetFlag(shader_data_registry.ShaderDataStandardFlagOutline)
		}
		for i := range e.Children {
			procChildren(EntityToStageEntity(e.Children[i]))
		}
	}
	procChildren(root)
}

func (m *StageManager) clearShaderDataFlag(root *StageEntity) {
	var procChildren func(e *StageEntity)
	procChildren = func(e *StageEntity) {
		if !m.IsSelected(e) {
			switch sd := e.StageData.ShaderData.(type) {
			case *shader_data_registry.ShaderDataStandard:
				sd.ClearFlag(shader_data_registry.ShaderDataStandardFlagOutline)
			case *shader_data_registry.ShaderDataStandardSkinned:
				sd.ClearFlag(shader_data_registry.ShaderDataStandardFlagOutline)
			case *shader_data_registry.ShaderDataPBR:
				sd.ClearFlag(shader_data_registry.ShaderDataStandardFlagOutline)
			}
		}
		for i := range e.Children {
			procChildren(EntityToStageEntity(e.Children[i]))
		}
	}
	procChildren(root)
}
