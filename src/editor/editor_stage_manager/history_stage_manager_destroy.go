package editor_stage_manager

import (
	"kaiju/platform/profiler/tracing"
)

type objectDeleteHistory struct {
	m *StageManager
	// TODO:  Only add the root-most entities to this list
	entities []*StageEntity
}

func (h *objectDeleteHistory) Redo() {
	defer tracing.NewRegion("objectDeleteHistory.Redo").End()
	for _, e := range h.entities {
		h.m.host.RemoveEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Deactivate()
		}
		h.m.OnEntityDestroy.Execute(e)
		e.isDeleted = true
	}
}

func (h *objectDeleteHistory) Undo() {
	defer tracing.NewRegion("objectDeleteHistory.Undo").End()
	for _, e := range h.entities {
		h.m.host.AddEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Activate()
		}
		h.m.OnEntitySpawn.Execute(e)
		e.isDeleted = false
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
		e.StageData.ShaderData.Destroy()
		e.Destroy()
	}
}
