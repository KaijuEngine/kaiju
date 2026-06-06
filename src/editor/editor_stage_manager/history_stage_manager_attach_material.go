/******************************************************************************/
/* history_stage_manager_attach_material.go                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import "log/slog"

type attachMaterialHistory struct {
	m         *StageManager
	e         *StageEntity
	fromMatId string
	toMatId   string
}

func (h *attachMaterialHistory) Redo() {
	if mat, err := h.m.host.MaterialCache().Material(h.toMatId); err == nil {
		h.e.SetMaterial(mat, h.m)
	} else {
		slog.Error("the material wasn't found in the cache to redo, which is unexpected",
			"id", mat.Id, "error", err)
	}
}

func (h *attachMaterialHistory) Undo() {
	if mat, err := h.m.host.MaterialCache().Material(h.fromMatId); err == nil {
		h.e.SetMaterial(mat, h.m)
	} else {
		slog.Error("the material wasn't found in the cache to redo, which is unexpected",
			"id", mat.Id, "error", err)
	}
}

func (h *attachMaterialHistory) Delete() {}
func (h *attachMaterialHistory) Exit()   {}
