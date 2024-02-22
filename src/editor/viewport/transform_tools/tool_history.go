package transform_tools

import (
	"kaiju/engine"
	"kaiju/matrix"
)

type toolHistory struct {
	entities []*engine.Entity
	from     []matrix.Vec3
	to       []matrix.Vec3
	state    ToolState
}

func (h *toolHistory) Redo() {
	for i, e := range h.entities {
		if h.state == ToolStateMove {
			e.Transform.SetPosition(h.to[i])
		} else if h.state == ToolStateRotate {
			e.Transform.SetRotation(h.to[i])
		} else if h.state == ToolStateScale {
			e.Transform.SetScale(h.to[i])
		}
	}
}

func (h *toolHistory) Undo() {
	for i, e := range h.entities {
		if h.state == ToolStateMove {
			e.Transform.SetPosition(h.from[i])
		} else if h.state == ToolStateRotate {
			e.Transform.SetRotation(h.from[i])
		} else if h.state == ToolStateScale {
			e.Transform.SetScale(h.from[i])
		}
	}
}

func (h *toolHistory) Delete() {}
func (h *toolHistory) Exit()   {}
