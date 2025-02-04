package ui

import (
	"kaiju/engine"
	"log/slog"
)

const (
	EntityDataName = "ui"
)

func FirstOnEntity(entity *engine.Entity) *UI {
	if entity == nil {
		slog.Error("the provided entity was nil")
		return nil
	}
	found := entity.NamedData(EntityDataName)
	if len(found) == 0 {
		return nil
	}
	return found[0].(*UI)
}

func FirstPanelOnEntity(entity *engine.Entity) *UI {
	ui := FirstOnEntity(entity)
	if ui == nil || ui.elmType == ElementTypeLabel {
		return nil
	}
	return ui
}

func AllOnEntity(entity *engine.Entity) []*UI {
	found := entity.NamedData(EntityDataName)
	if len(found) == 0 {
		return []*UI{}
	}
	res := make([]*UI, len(found))
	for i := range found {
		res[i] = found[i].(*UI)
	}
	return res
}

func FindByName(list []*UI, name string) *UI {
	var found *UI
	for i := 0; i < len(list) && found == nil; i++ {
		found = list[i].FindByName(name)
	}
	return found
}

func DestroyList(list []*UI) {
	for i := range list {
		list[i].entity.Destroy()
	}
}
