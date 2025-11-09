package stage_workspace

import (
	"kaiju/editor/editor_stage_manager"
)

type detailSetNameHistory struct {
	w        *WorkspaceDetailsUI
	entities []*editor_stage_manager.StageEntity
	prevName []string
	nextName string
}

func (h *detailSetNameHistory) Redo() {
	w := h.w.workspace.Value()
	for _, e := range h.entities {
		e.SetName(h.nextName)
		w.hierarchyUI.updateEntityName(e.StageData.Description.Id, h.nextName)
	}
	h.w.detailsName.UI.ToInput().SetTextWithoutEvent(h.nextName)
}

func (h *detailSetNameHistory) Undo() {
	w := h.w.workspace.Value()
	for i, e := range h.entities {
		e.SetName(h.prevName[i])
		w.hierarchyUI.updateEntityName(e.StageData.Description.Id, h.prevName[i])
	}
	h.w.detailsName.UI.ToInput().SetTextWithoutEvent(h.prevName[0])
}

func (h *detailSetNameHistory) Delete() {}
func (h *detailSetNameHistory) Exit()   {}
