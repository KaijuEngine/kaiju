package editor_stage_manager

import (
	"kaiju/platform/profiler/tracing"
)

type objectSpawnHistory struct {
	m *StageManager
	e *StageEntity
}

func (h *objectSpawnHistory) Redo() {
	defer tracing.NewRegion("objectSpawnHistory.Redo").End()
	h.m.host.AddEntity(&h.e.Entity)
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Activate()
	}
	h.m.OnEntitySpawn.Execute(h.e)
}

func (h *objectSpawnHistory) Undo() {
	defer tracing.NewRegion("objectSpawnHistory.Undo").End()
	h.m.host.RemoveEntity(&h.e.Entity)
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Deactivate()
	}
	h.m.OnEntityDestroy.Execute(h.e)
}

func (h *objectSpawnHistory) Delete() {
	h.e.StageData.ShaderData.Destroy()
	h.e.Destroy()
}

func (h *objectSpawnHistory) Exit() {}
