/******************************************************************************/
/* history_stage_workspace_material_change.go                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import "kaijuengine.com/editor/editor_stage_manager"

type detailsMaterialChangeHistory struct {
	workspace      *StageWorkspace
	detailsUI      *WorkspaceDetailsUI
	entity         *editor_stage_manager.StageEntity
	fromMaterialId string
	toMaterialId   string
	fromTextureIds []string
	toTextureIds   []string
}

func (h *detailsMaterialChangeHistory) apply(materialId string, textureIds []string) bool {
	if h.workspace == nil || h.entity == nil {
		return false
	}
	if !h.workspace.setEntityMaterial(h.entity, materialId, textureIds) {
		return false
	}
	if h.detailsUI != nil {
		h.detailsUI.setMaterialInputValue(materialId)
		h.detailsUI.reload()
	}
	return true
}

func (h *detailsMaterialChangeHistory) Redo() {
	h.apply(h.toMaterialId, h.toTextureIds)
}

func (h *detailsMaterialChangeHistory) Undo() {
	h.apply(h.fromMaterialId, h.fromTextureIds)
}

func (h *detailsMaterialChangeHistory) Delete() {}
func (h *detailsMaterialChangeHistory) Exit()   {}
