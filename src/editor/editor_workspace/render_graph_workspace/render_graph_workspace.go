/******************************************************************************/
/* render_graph_workspace.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	ID          = "renderGraph"
	DisplayName = "Render Graph"
)

func init() {
	editor_workspace_registry.Register(&RenderGraphWorkspace{})
}

type RenderGraphWorkspace struct {
	common_workspace.CommonWorkspace
	ed              editor_workspace.WorkspaceEditorInterface
	stageView       *editor_stage_view.StageView
	root            *document.Element
	stageViewport   *document.Element
	shaderGraphArea *document.Element
	dimensionToggle *document.Element
	graph           shaderGraph
	createNodeMenu  shaderGraphCreateNodeMenu
	createNodeCount int
}

type RenderGraphWorkspaceUIData struct {
	CameraMode  string
	CreateNodes []shaderGraphNodeMenuData
}

func (w *RenderGraphWorkspace) ID() string          { return ID }
func (w *RenderGraphWorkspace) DisplayName() string { return DisplayName }
func (w *RenderGraphWorkspace) IsRequired() bool    { return false }

func (w *RenderGraphWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("RenderGraphWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	data := RenderGraphWorkspaceUIData{
		CameraMode:  w.stageView.Camera().ModeString(),
		CreateNodes: shaderGraphNodeCatalogMenuData(),
	}
	if err := w.CommonWorkspace.InitializeWithUI(ed.Host(),
		"editor/ui/workspace/render_graph_workspace.go.html", data, map[string]func(*document.Element){
			"toggleDimension":      w.toggleDimension,
			"filterCreateNodeMenu": w.filterCreateNodeMenu,
			"selectCreateNode":     w.selectCreateNode,
			"closeCreateNodeMenu":  w.closeCreateNodeMenu,
		}); err != nil {
		return err
	}
	w.root, _ = w.Doc.GetElementById("renderGraphWorkspace")
	w.stageViewport, _ = w.Doc.GetElementById("stageViewport")
	w.shaderGraphArea, _ = w.Doc.GetElementById("shaderGraphArea")
	w.dimensionToggle, _ = w.Doc.GetElementById("dimensionToggle")
	if w.root != nil {
		w.root.UIPanel.AllowClickThrough()
	}
	w.createNodeMenu.Initialize(w)
	w.graph.Initialize(ed.Host())
	source := w.graph.CreateNode(shaderGraphNodeSpec{
		Name:        "Test Node",
		Description: "Temporary shader graph node used to verify layout, clipping, and dragging.",
		Inputs: []shaderGraphPortSpec{
			{Name: "Base Color", Type: "color"},
			{Name: "Roughness", Type: "float"},
			{Name: "Normal", Type: "vec3"},
		},
		Outputs: []shaderGraphPortSpec{
			{Name: "Material", Type: "surface"},
		},
	}, matrix.NewVec2(42, 56))
	output := w.graph.CreateNode(shaderGraphNodeSpec{
		Name:        "Material Output",
		Description: "Temporary output node used to verify connected shader graph sockets.",
		Inputs: []shaderGraphPortSpec{
			{Name: "Surface", Type: "surface"},
			{Name: "Volume", Type: "volume"},
			{Name: "Displacement", Type: "vec3"},
		},
	}, matrix.NewVec2(350, 150))
	if source != nil && output != nil {
		w.graph.CreateConnection(source.Output(0), output.Input(0))
	}
	return nil
}

func (w *RenderGraphWorkspace) Shutdown() {
	defer tracing.NewRegion("RenderGraphWorkspace.Shutdown").End()
	w.graph.Shutdown()
	w.CommonShutdown()
}

func (w *RenderGraphWorkspace) Open() {
	defer tracing.NewRegion("RenderGraphWorkspace.Open").End()
	w.CommonOpen()
	w.applyLayout()
	if w.stageViewport != nil {
		w.stageView.SetViewportUI(w.stageViewport.UI)
	}
	if w.dimensionToggle != nil {
		w.dimensionToggle.InnerLabel().SetText(w.stageView.Camera().ModeString())
	}
	w.createNodeMenu.Hide()
	w.graph.Open()
	w.stageView.Open()
}

func (w *RenderGraphWorkspace) Close() {
	defer tracing.NewRegion("RenderGraphWorkspace.Close").End()
	if w.stageView != nil {
		w.stageView.SetViewportUI(nil)
		w.stageView.Close()
	}
	w.graph.Close()
	w.createNodeMenu.Hide()
	w.CommonClose()
}

func (w *RenderGraphWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *RenderGraphWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("RenderGraphWorkspace.Update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred {
		return
	}
	w.applyLayout()
	w.createNodeMenu.Update()
	w.graph.Update()
	if w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *RenderGraphWorkspace) toggleDimension(e *document.Element) {
	defer tracing.NewRegion("RenderGraphWorkspace.toggleDimension").End()
	lbl := e.InnerLabel()
	switch lbl.Text() {
	case "3D":
		lbl.SetText("2D")
		w.stageView.SetCameraMode(editor_controls.EditorCameraMode2d)
	case "2D":
		lbl.SetText("3D")
		w.stageView.SetCameraMode(editor_controls.EditorCameraMode3d)
	}
}

func (w *RenderGraphWorkspace) ShowCreateNodeMenu() {
	w.applyLayout()
	position := w.createNodeMenuPosition()
	w.createNodeMenu.Show(position, w.graph.graphPositionFromView(position))
}

func (w *RenderGraphWorkspace) CreateNodeFromAction(args CreateNodeActionArgs) (*shaderGraphNode, bool) {
	spec, ok := shaderGraphNodeCatalogSpec(args.NodeID)
	if !ok {
		return nil, false
	}
	position := args.position(w.defaultCreateNodePosition())
	node := w.graph.CreateNode(spec, position)
	if node == nil {
		return nil, false
	}
	w.createNodeCount++
	w.createNodeMenu.Hide()
	return node, true
}

func (w *RenderGraphWorkspace) runCreateNodeAction(nodeID string) {
	position := w.createNodeMenu.CreatePosition()
	w.ed.Actions().Run(editor_action.Request{
		ID: ActionRenderGraphCreateNode,
		Params: CreateNodeActionArgs{
			NodeID:      nodeID,
			X:           float32(position.X()),
			Y:           float32(position.Y()),
			UsePosition: true,
		},
		Source: editor_action.SourceMenu,
	})
}
