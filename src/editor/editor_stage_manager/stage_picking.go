/******************************************************************************/
/* stage_picking.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"log/slog"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const editorPickingShaderDataName = "editor_pick"

func (m *StageManager) AssignPickID(e *StageEntity) uint32 {
	defer tracing.NewRegion("StageManager.AssignPickID").End()
	if e == nil {
		return 0
	}
	if m.pickIDToEntity == nil {
		m.pickIDToEntity = make(map[uint32]*StageEntity)
	}
	if e.PickID != 0 {
		if current := m.pickIDToEntity[e.PickID]; current == nil || current == e {
			m.pickIDToEntity[e.PickID] = e
			if e.PickID > m.nextPickID {
				m.nextPickID = e.PickID
			}
			return e.PickID
		}
		e.PickID = 0
	}
	for {
		m.nextPickID++
		if m.nextPickID == 0 {
			continue
		}
		if _, exists := m.pickIDToEntity[m.nextPickID]; exists {
			continue
		}
		e.PickID = m.nextPickID
		m.pickIDToEntity[e.PickID] = e
		return e.PickID
	}
}

func (m *StageManager) EntityByPickID(id uint32) (*StageEntity, bool) {
	defer tracing.NewRegion("StageManager.EntityByPickID").End()
	if id == 0 || m.pickIDToEntity == nil {
		return nil, false
	}
	e := m.pickIDToEntity[id]
	if e == nil || e.IsDeleted() || e.IsLocked() {
		return nil, false
	}
	return e, true
}

func (m *StageManager) unregisterPickID(e *StageEntity) {
	if e == nil || e.PickID == 0 || m.pickIDToEntity == nil {
		return
	}
	if m.pickIDToEntity[e.PickID] == e {
		delete(m.pickIDToEntity, e.PickID)
	}
}

func (m *StageManager) newPickingDrawing(e *StageEntity, material *rendering.Material) (rendering.Drawing, bool) {
	if e == nil || material == nil || e.StageData.Mesh == nil || e.StageData.PickingShaderData != nil {
		return rendering.Drawing{}, false
	}
	pickID := m.AssignPickID(e)
	sd := shader_data_registry.Create(editorPickingShaderDataName)
	sd.(*shader_data_registry.ShaderDataEditorPicking).PickID = pickID
	e.StageData.PickingShaderData = sd
	var culler rendering.ViewCuller
	if m.host != nil {
		culler = &m.host.Cameras.Primary
	}
	return rendering.Drawing{
		Material:   material,
		Mesh:       e.StageData.Mesh,
		ShaderData: sd,
		Transform:  &e.Transform,
		ViewCuller: culler,
		Layer:      rendering.RenderLayerEditorPicking,
	}, true
}

func (m *StageManager) addPickingDrawing(e *StageEntity) {
	defer tracing.NewRegion("StageManager.addPickingDrawing").End()
	if m.host == nil {
		return
	}
	material, err := m.host.MaterialCache().Material(assets.MaterialDefinitionEditorPicking)
	if err != nil {
		slog.Warn("failed to create editor picking material", "error", err)
		return
	}
	if draw, ok := m.newPickingDrawing(e, material); ok {
		m.host.Drawings.AddDrawing(draw)
	}
}
