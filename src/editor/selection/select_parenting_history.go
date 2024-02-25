package selection

import "kaiju/engine"

type selectParentingHistory struct {
	targets     []*engine.Entity
	lastParents []*engine.Entity
	newParent   *engine.Entity
}

func (h *selectParentingHistory) Redo() {
	for _, e := range h.targets {
		e.SetParent(h.newParent)
	}
}

func (h *selectParentingHistory) Undo() {
	for i, e := range h.targets {
		e.SetParent(h.lastParents[i])
	}
}

func (h *selectParentingHistory) Delete() {}
func (h *selectParentingHistory) Exit()   {}
