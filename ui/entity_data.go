package ui

import "kaiju/engine"

const (
	EntityDataName = "ui"
)

func FirstOnEntity(entity *engine.Entity) UI {
	found := entity.NamedData(EntityDataName)
	if len(found) == 0 {
		return nil
	}
	return found[0].(UI)
}

func AllOnEntity(entity *engine.Entity) []UI {
	found := entity.NamedData(EntityDataName)
	if len(found) == 0 {
		return []UI{}
	}
	res := make([]UI, len(found))
	for i := range found {
		res[i] = found[i].(UI)
	}
	return res
}
