package editor_stage_manager

import (
	"kaiju/platform/profiler/tracing"
	"unsafe"
)

type objectSpawnHistory struct {
	m *StageManager
	e *StageEntity
}

func (h *objectSpawnHistory) Redo() {
	defer tracing.NewRegion("objectSpawnHistory.Redo").End()
	var recurseAddEntity func(e *StageEntity)
	recurseAddEntity = func(e *StageEntity) {
		h.m.host.AddEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Activate()
		}
		h.m.OnEntitySpawn.Execute(e)
		for i := range e.Children {
			recurseAddEntity((*StageEntity)(unsafe.Pointer(e.Children[i])))
		}
	}
	recurseAddEntity(h.e)
}

func (h *objectSpawnHistory) Undo() {
	defer tracing.NewRegion("objectSpawnHistory.Undo").End()
	var recurseRemoveEntity func(e *StageEntity)
	recurseRemoveEntity = func(e *StageEntity) {
		h.m.host.RemoveEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Deactivate()
		}
		h.m.OnEntityDestroy.Execute(e)
		for i := range e.Children {
			recurseRemoveEntity((*StageEntity)(unsafe.Pointer(e.Children[i])))
		}
	}
	recurseRemoveEntity(h.e)
}

func (h *objectSpawnHistory) Delete() {
	h.e.StageData.ShaderData.Destroy()
	h.e.Destroy()
}

func (h *objectSpawnHistory) Exit() {}
