/******************************************************************************/
/* history_stage_manager_destroy.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import "kaijuengine.com/platform/profiler/tracing"

type objectDeleteHistory struct {
	m *StageManager
	// TODO:  Only add the root-most entities to this list
	entities []*StageEntity
}

func (h *objectDeleteHistory) Redo() {
	defer tracing.NewRegion("objectDeleteHistory.Redo").End()
	for _, e := range h.entities {
		e.Deactivate()
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Deactivate()
		}
		if e.StageData.PickingShaderData != nil {
			e.StageData.PickingShaderData.Deactivate()
		}
		h.m.OnEntityDestroy.Execute(e)
		e.isDeleted = true
		if e.StageData.Bvh != nil {
			h.m.RemoveEntityBVH(e)
		}
	}
}

func (h *objectDeleteHistory) Undo() {
	defer tracing.NewRegion("objectDeleteHistory.Undo").End()
	for _, e := range h.entities {
		e.Activate()
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Activate()
		}
		if e.StageData.PickingShaderData != nil {
			e.StageData.PickingShaderData.Activate()
		}
		h.m.OnEntitySpawn.Execute(e)
		e.isDeleted = false
		if e.StageData.Bvh != nil {
			h.m.AddBVH(e)
		}
	}
	for _, e := range h.entities {
		if e.Parent != nil {
			h.m.OnEntityChangedParent.Execute(e)
		}
	}
}

func (h *objectDeleteHistory) Delete() {}

func (h *objectDeleteHistory) Exit() {
	for _, e := range h.entities {
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Destroy()
		}
		if e.StageData.PickingShaderData != nil {
			e.StageData.PickingShaderData.Destroy()
		}
		h.m.host.DestroyEntity(&e.Entity)
	}
}
