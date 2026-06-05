/******************************************************************************/
/* render_graph_workspace.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"log/slog"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_overlay/content_selector"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui/markup/document"
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
	sidePanel       *document.Element
	stageViewport   *document.Element
	shaderGraphArea *document.Element
	dimensionToggle *document.Element
	nameInput       *document.Element
	status          *document.Element
	graph           shaderGraph
	createNodeMenu  shaderGraphCreateNodeMenu
	createNodeCount int
	currentGraphID  string
	currentName     string
	generated       RenderGraphGenerated
}

type RenderGraphWorkspaceUIData struct {
	CameraMode  string
	CreateNodes []shaderGraphNodeMenuData
	GraphName   string
}

func (w *RenderGraphWorkspace) ID() string          { return ID }
func (w *RenderGraphWorkspace) DisplayName() string { return DisplayName }
func (w *RenderGraphWorkspace) IsRequired() bool    { return false }

func (w *RenderGraphWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("RenderGraphWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	w.currentName = defaultRenderGraphName
	data := RenderGraphWorkspaceUIData{
		CameraMode:  w.stageView.Camera().ModeString(),
		CreateNodes: shaderGraphNodeCatalogMenuData(),
		GraphName:   w.currentName,
	}
	if err := w.CommonWorkspace.InitializeWithUI(ed.Host(),
		"editor/ui/workspace/render_graph_workspace.go.html", data, map[string]func(*document.Element){
			"toggleDimension":      w.toggleDimension,
			"filterCreateNodeMenu": w.filterCreateNodeMenu,
			"selectCreateNode":     w.selectCreateNode,
			"closeCreateNodeMenu":  w.closeCreateNodeMenu,
			"renameRenderGraph":    w.renameRenderGraph,
			"newRenderGraph":       w.newRenderGraph,
			"loadRenderGraph":      w.loadRenderGraph,
			"saveRenderGraph":      w.saveRenderGraph,
		}); err != nil {
		return err
	}
	w.root, _ = w.Doc.GetElementById("renderGraphWorkspace")
	w.sidePanel, _ = w.Doc.GetElementById("renderGraphPanel")
	w.stageViewport, _ = w.Doc.GetElementById("stageViewport")
	w.shaderGraphArea, _ = w.Doc.GetElementById("shaderGraphArea")
	w.dimensionToggle, _ = w.Doc.GetElementById("dimensionToggle")
	w.nameInput, _ = w.Doc.GetElementById("renderGraphName")
	w.status, _ = w.Doc.GetElementById("renderGraphStatus")
	if w.root != nil {
		w.root.UIPanel.AllowClickThrough()
	}
	w.createNodeMenu.Initialize(w)
	w.graph.Initialize(ed.Host(), ed.History())
	w.graph.zoomBlocked = w.createNodeMenu.BlocksGraphZoom
	w.graph.inputBlocked = w.createNodeMenu.BlocksGraphInput
	w.graph.selectTexture = w.selectGraphTexture
	w.graph.textureName = w.graphTextureName
	w.resetGraphToDefault()
	w.updateGraphNameInput()
	w.setRenderGraphStatus("Unsaved render graph")
	return nil
}

func (w *RenderGraphWorkspace) Shutdown() {
	defer tracing.NewRegion("RenderGraphWorkspace.Shutdown").End()
	if w.stageView != nil {
		w.stageView.ClearViewportToolOwner(w)
	}
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
	w.stageView.SetViewportToolOwner(w)
}

func (w *RenderGraphWorkspace) Close() {
	defer tracing.NewRegion("RenderGraphWorkspace.Close").End()
	if w.stageView != nil {
		w.stageView.ClearViewportToolOwner(w)
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

func (w *RenderGraphWorkspace) selectGraphTexture(current string, onSelect func(string), onClose func()) {
	if w.ed == nil || w.Host == nil || w.ed.Cache() == nil {
		return
	}
	w.ed.BlurInterface()
	if _, err := content_selector.Show(w.Host, (content_database.Texture{}).TypeName(), w.ed.Cache(), func(id string) {
		w.ed.FocusInterface()
		if onSelect != nil {
			onSelect(id)
		}
	}, func() {
		w.ed.FocusInterface()
		if onClose != nil {
			onClose()
		}
	}); err != nil {
		w.ed.FocusInterface()
		slog.Error("failed to show render graph texture selector", "texture", current, "error", err)
	}
}

func (w *RenderGraphWorkspace) graphTextureName(id string) string {
	if id == "" {
		return "None"
	}
	if id == assets.TextureSquare {
		return assets.TextureSquare
	}
	if w.ed == nil || w.ed.Cache() == nil {
		return id
	}
	cc, err := w.ed.Cache().Read(id)
	if err != nil || cc.Config.Name == "" {
		return id
	}
	return cc.Config.Name
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

func (w *RenderGraphWorkspace) CenterView() {
	w.graph.CenterView()
}

func (w *RenderGraphWorkspace) FocusSelectedNodes() bool {
	w.applyLayout()
	return w.graph.FocusSelection()
}

func (w *RenderGraphWorkspace) CreateNodeFromAction(args CreateNodeActionArgs) (*shaderGraphNode, bool) {
	if _, ok := shaderGraphNodeCatalogSpec(args.NodeID); !ok {
		return nil, false
	}
	position := args.position(w.defaultCreateNodePosition())
	previousSelection := w.graph.selectionIDs()
	node, ok := w.graph.CreateCatalogNode(args.NodeID, position)
	if !ok || node == nil {
		return nil, false
	}
	w.graph.cancelBoxSelection()
	w.graph.setSelectionNodes([]*shaderGraphNode{node})
	if w.graph.history != nil {
		w.graph.history.Add(&shaderGraphNodeCreateHistory{
			graph:             &w.graph,
			node:              renderGraphNodeFromShaderGraphNode(node),
			previousSelection: previousSelection,
		})
	}
	w.createNodeCount++
	w.createNodeMenu.Hide()
	return node, true
}

func (w *RenderGraphWorkspace) GraphDocument() RenderGraphDocument {
	document := w.graph.Document()
	document.Name = w.currentName
	if !w.generated.IsZero() {
		generated := w.generated
		document.Generated = &generated
	}
	return document
}

func (w *RenderGraphWorkspace) SerializeGraph() ([]byte, error) {
	return SerializeRenderGraphDocument(w.GraphDocument())
}

func (w *RenderGraphWorkspace) DeserializeGraph(data []byte) error {
	document, err := DeserializeRenderGraphDocument(data)
	if err != nil {
		return err
	}
	if err = w.graph.LoadDocument(document); err != nil {
		return err
	}
	w.currentGraphID = ""
	w.generated = RenderGraphGenerated{}
	if document.Generated != nil {
		w.generated = *document.Generated
	}
	if document.Name != "" {
		w.currentName = document.Name
	}
	w.updateGraphNameInput()
	return nil
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
