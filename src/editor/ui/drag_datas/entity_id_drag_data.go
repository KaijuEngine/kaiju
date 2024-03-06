package drag_datas

import "kaiju/engine"

type EntityIdDragData struct {
	EntityId engine.EntityId
}

func (e *EntityIdDragData) DragUpdate() {
}
