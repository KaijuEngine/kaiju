/******************************************************************************/
/* history_stage_manager_spawn.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import "kaijuengine.com/platform/profiler/tracing"

type objectSpawnHistory struct {
	m *StageManager
	e *StageEntity
}

func (h *objectSpawnHistory) Redo() {
	defer tracing.NewRegion("objectSpawnHistory.Redo").End()
	h.e.Activate()
	h.e.isDeleted = false
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Activate()
	}
	if h.e.StageData.PickingShaderData != nil {
		h.e.StageData.PickingShaderData.Activate()
	}
	if h.e.StageData.Bvh != nil {
		h.m.AddBVH(h.e)
	}
	h.m.OnEntitySpawn.Execute(h.e)
}

func (h *objectSpawnHistory) Undo() {
	defer tracing.NewRegion("objectSpawnHistory.Undo").End()
	h.e.Deactivate()
	h.e.isDeleted = true
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Deactivate()
	}
	if h.e.StageData.PickingShaderData != nil {
		h.e.StageData.PickingShaderData.Deactivate()
	}
	if h.e.StageData.Bvh != nil {
		h.m.RemoveEntityBVH(h.e)
	}
	h.m.OnEntityDestroy.Execute(h.e)
}

func (h *objectSpawnHistory) Delete() {
	h.m.host.DestroyEntity(&h.e.Entity)
	h.e.ForceCleanup()
}

func (h *objectSpawnHistory) Exit() {}
