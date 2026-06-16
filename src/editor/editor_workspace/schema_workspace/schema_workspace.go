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
	"kaijuengine.com/engine/ui/markup/document"
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
	ed     editor_workspace.WorkspaceEditorInterface
	graph  schemaGraph
	canvas *document.Element
}

func (w *SchemaWorkspace) ID() string          { return ID }
func (w *SchemaWorkspace) DisplayName() string { return DisplayName }
func (w *SchemaWorkspace) IsRequired() bool    { return false }

func (w *SchemaWorkspace) IsFocusedOnInput() bool {
	return w.CommonWorkspace.IsFocusedOnInput() || w.graph.IsFocusedOnInput()
}

func (w *SchemaWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("SchemaWorkspace.Initialize").End()
	w.ed = ed
	if err := w.CommonWorkspace.InitializeWithUI(ed.Host(), uiFile, nil, map[string]func(*document.Element){
		"clickAddProperties":  w.clickAddProperties,
		"clickAddDefinitions": w.clickAddDefinitions,
	}); err != nil {
		return err
	}
	w.canvas, _ = w.Doc.GetElementById("schemaCanvas")
	w.graph.Initialize(ed.Host())
	w.applyLayout()
	return nil
}

func (w *SchemaWorkspace) Shutdown() {
	defer tracing.NewRegion("SchemaWorkspace.Shutdown").End()
	w.graph.Shutdown()
	w.canvas = nil
	w.CommonShutdown()
}

func (w *SchemaWorkspace) Open() {
	defer tracing.NewRegion("SchemaWorkspace.Open").End()
	w.CommonOpen()
	w.applyLayout()
	w.graph.Open()
}

func (w *SchemaWorkspace) Close() {
	defer tracing.NewRegion("SchemaWorkspace.Close").End()
	w.graph.Close()
	w.CommonClose()
}

func (w *SchemaWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *SchemaWorkspace) Update(float64) {
	defer tracing.NewRegion("SchemaWorkspace.Update").End()
	if w.UiMan.IsUpdateDisabled() || w.IsBlurred {
		return
	}
	w.applyLayout()
	w.graph.Update()
}

func (w *SchemaWorkspace) AddPropertiesNode() {
	defer tracing.NewRegion("SchemaWorkspace.AddPropertiesNode").End()
	w.graph.CreateRootNode(schemaNodeKindProperties)
}

func (w *SchemaWorkspace) AddDefinitionsNode() {
	defer tracing.NewRegion("SchemaWorkspace.AddDefinitionsNode").End()
	w.graph.CreateRootNode(schemaNodeKindDefinitions)
}

func (w *SchemaWorkspace) NodeCount() int {
	return w.graph.NodeCount()
}

func (w *SchemaWorkspace) clickAddProperties(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.clickAddProperties").End()
	w.AddPropertiesNode()
}

func (w *SchemaWorkspace) clickAddDefinitions(*document.Element) {
	defer tracing.NewRegion("SchemaWorkspace.clickAddDefinitions").End()
	w.AddDefinitionsNode()
}
