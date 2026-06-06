/******************************************************************************/
/* history_stage_workspace_mesh_change.go                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import "kaijuengine.com/editor/editor_stage_manager"

type detailsMeshChangeHistory struct {
	workspace  *StageWorkspace
	detailsUI  *WorkspaceDetailsUI
	entity     *editor_stage_manager.StageEntity
	fromMeshId string
	toMeshId   string
	fromMatId  string
	toMatId    string
	fromTexIds []string
	toTexIds   []string
	toStateSet bool
}

func (h *detailsMeshChangeHistory) apply(meshId string) bool {
	if h.workspace == nil || h.entity == nil {
		return false
	}
	var ok bool
	if meshId == "" {
		ok = h.workspace.clearEntityMesh(h.entity)
	} else {
		ok = h.workspace.setEntityMesh(h.entity, meshId)
	}
	if !ok {
		return false
	}
	switch meshId {
	case h.fromMeshId:
		if !h.workspace.setEntityMaterial(h.entity, h.fromMatId, h.fromTexIds) {
			return false
		}
	case h.toMeshId:
		if !h.toStateSet {
			break
		}
		if !h.workspace.setEntityMaterial(h.entity, h.toMatId, h.toTexIds) {
			return false
		}
	}
	if h.detailsUI != nil {
		h.detailsUI.setMeshInputValue(meshId)
		h.detailsUI.reload()
	}
	return true
}

func (h *detailsMeshChangeHistory) Redo() { h.apply(h.toMeshId) }

func (h *detailsMeshChangeHistory) Undo() { h.apply(h.fromMeshId) }

func (h *detailsMeshChangeHistory) Delete() {}
func (h *detailsMeshChangeHistory) Exit()   {}
