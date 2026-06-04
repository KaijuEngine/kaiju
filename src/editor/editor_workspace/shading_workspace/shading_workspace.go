/******************************************************************************/
/* shading_workspace.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import (
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
)

const (
	ID          = "shading"
	DisplayName = "Shading"
)

func init() {
	editor_workspace_registry.Register(&ShadingWorkspace{})
}

type ShadingWorkspace struct {
	common_workspace.CommonWorkspace
}

func (w *ShadingWorkspace) ID() string          { return ID }
func (w *ShadingWorkspace) DisplayName() string { return DisplayName }
func (w *ShadingWorkspace) IsRequired() bool    { return false }

func (w *ShadingWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	return w.CommonWorkspace.InitializeWithUI(ed.Host(),
		"editor/ui/workspace/shading_workspace.go.html", nil, nil)
}

func (w *ShadingWorkspace) Shutdown() {
	w.CommonShutdown()
}

func (w *ShadingWorkspace) Open() {
	w.CommonOpen()
}

func (w *ShadingWorkspace) Close() {
	w.CommonClose()
}

func (w *ShadingWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}
