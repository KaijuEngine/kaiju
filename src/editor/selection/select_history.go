package selection

import "kaiju/engine"

type selectHistory struct {
	selection *Selection
	from      []*engine.Entity
	to        []*engine.Entity
	// TODO:  Pointers won't due when we delete/restore entities
	// we'll likely want to not actually destroy the entity until
	// it falls off the history stack as that greatly simplifies
	// code like this
}

func (h *selectHistory) Redo() {
	h.selection.setInternal(h.to)
}

func (h *selectHistory) Undo() {
	h.selection.setInternal(h.from)
}

func (h *selectHistory) Delete() {}
func (h *selectHistory) Exit()   {}
