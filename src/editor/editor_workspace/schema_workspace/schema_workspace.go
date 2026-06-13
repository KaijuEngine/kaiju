/******************************************************************************/
/* schema_workspace.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import (
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	ID          = "schema"
	DisplayName = "Schema"

	uiFile = "editor/ui/workspace/schema_workspace.go.html"
)

func init() {
	editor_workspace_registry.Register(&SchemaWorkspace{})
}

type SchemaWorkspace struct {
	common_workspace.CommonWorkspace
}

func (w *SchemaWorkspace) ID() string          { return ID }
func (w *SchemaWorkspace) DisplayName() string { return DisplayName }
func (w *SchemaWorkspace) IsRequired() bool    { return false }

func (w *SchemaWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("SchemaWorkspace.Initialize").End()
	return w.CommonWorkspace.InitializeWithUI(ed.Host(), uiFile, nil, nil)
}

func (w *SchemaWorkspace) Shutdown() {
	defer tracing.NewRegion("SchemaWorkspace.Shutdown").End()
	w.CommonShutdown()
}

func (w *SchemaWorkspace) Open() {
	defer tracing.NewRegion("SchemaWorkspace.Open").End()
	w.CommonOpen()
}

func (w *SchemaWorkspace) Close() {
	defer tracing.NewRegion("SchemaWorkspace.Close").End()
	w.CommonClose()
}

func (w *SchemaWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}
