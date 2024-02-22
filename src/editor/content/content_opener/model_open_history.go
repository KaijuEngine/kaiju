package content_opener

import (
	"kaiju/engine"
	"kaiju/rendering"
)

type modelOpenHistory struct {
	host     *engine.Host
	entity   *engine.Entity
	drawings []rendering.Drawing
}

func (h *modelOpenHistory) Redo() {
	h.entity.Activate()
	for i := range h.drawings {
		h.drawings[i].ShaderData.Activate()
	}
	h.host.AddEntity(h.entity)
}

func (h *modelOpenHistory) Undo() {
	h.entity.Deactivate()
	for i := range h.drawings {
		h.drawings[i].ShaderData.Deactivate()
	}
	h.host.RemoveEntity(h.entity)
}

func (h *modelOpenHistory) Delete() {
	h.entity.Destroy()
	for i := range h.drawings {
		h.drawings[i].ShaderData.Destroy()
	}
}

func (h *modelOpenHistory) Exit() {}
