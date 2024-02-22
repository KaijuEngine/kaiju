package deleter

import (
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/rendering"
)

type deleteHistory struct {
	entities  []*engine.Entity
	selection *selection.Selection
}

func (h *deleteHistory) Redo() {
	for _, e := range h.entities {
		for _, d := range e.NamedData("drawing") {
			d.(*rendering.Drawing).ShaderData.Deactivate()
		}
		e.Deactivate()
	}
	if h.selection != nil {
		h.selection.UntrackedClear()
	}
}

func (h *deleteHistory) Undo() {
	for _, e := range h.entities {
		for _, d := range e.NamedData("drawing") {
			d.(*rendering.Drawing).ShaderData.Activate()
		}
		e.Activate()
	}
	if h.selection != nil {
		h.selection.UntrackedAdd(h.entities...)
	}
}

func (h *deleteHistory) Delete() {}

func (h *deleteHistory) Exit() {
	for _, e := range h.entities {
		for _, d := range e.NamedData("drawing") {
			d.(*rendering.Drawing).ShaderData.Destroy()
		}
		e.Destroy()
	}
}
