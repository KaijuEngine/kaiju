package editor_stage_manager

import (
	"kaiju/platform/profiler/tracing"
	"unsafe"
)

type objectDeleteHistory struct {
	m *StageManager
	// TODO:  Only add the root-most entities to this list
	entities []*StageEntity
}

func (h *objectDeleteHistory) Redo() {
	defer tracing.NewRegion("objectDeleteHistory.Redo").End()
	// TODO:  When deleting a selection of entities, it should do:
	// 1. Select only the root-most parents
	// 2. Fake-delete each of them
	// 3. Clear the selection, without it being added to history
	var recurseRemoveEntity func(e *StageEntity)
	recurseRemoveEntity = func(e *StageEntity) {
		h.m.host.RemoveEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Deactivate()
		}
		for i := range e.Children {
			recurseRemoveEntity((*StageEntity)(unsafe.Pointer(e.Children[i])))
		}
	}
	for _, e := range h.entities {
		recurseRemoveEntity(e)
	}
}

func (h *objectDeleteHistory) Undo() {
	defer tracing.NewRegion("objectDeleteHistory.Undo").End()
	var recurseAddEntity func(e *StageEntity)
	recurseAddEntity = func(e *StageEntity) {
		h.m.host.AddEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Activate()
		}
		for i := range e.Children {
			recurseAddEntity((*StageEntity)(unsafe.Pointer(e.Children[i])))
		}
	}
	for _, e := range h.entities {
		recurseAddEntity(e)
	}
}

func (h *objectDeleteHistory) Delete() {}

func (h *objectDeleteHistory) Exit() {
	for _, e := range h.entities {
		e.StageData.ShaderData.Destroy()
		e.Destroy()
	}
}
