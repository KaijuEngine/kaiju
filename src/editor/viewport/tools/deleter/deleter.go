package deleter

import (
	"kaiju/editor/memento"
	"kaiju/editor/selection"
	"kaiju/engine"
)

func doDelete(h *deleteHistory, history *memento.History) {
	h.Redo()
	history.Add(h)
}

func Delete(history *memento.History, entities []*engine.Entity) {
	h := &deleteHistory{entities, nil}
	doDelete(h, history)
}

func DeleteSelected(history *memento.History,
	selection *selection.Selection, entities []*engine.Entity) {

	h := &deleteHistory{entities, selection}
	doDelete(h, history)
}
