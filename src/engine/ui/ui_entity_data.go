/******************************************************************************/
/* ui_entity_data.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "kaijuengine.com/engine"

const (
	EntityDataName = "ui"
)

func FirstOnEntity(entity *engine.Entity) *UI {
	if entity == nil {
		return nil
	}
	found := entity.NamedData(EntityDataName)
	if len(found) == 0 {
		return nil
	}
	return found[0].(*UI)
}

func FirstPanelOnEntity(entity *engine.Entity) *Panel {
	ui := FirstOnEntity(entity)
	if ui == nil {
		return nil
	}
	return (*Panel)(ui)
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
