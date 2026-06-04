/******************************************************************************/
/* shading_workspace.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import (
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
	ID          = "shading"
	DisplayName = "Shading"
)

func init() {
	editor_workspace_registry.Register(&ShadingWorkspace{})
}

type ShadingWorkspace struct {
	common_workspace.CommonWorkspace
	ed              editor_workspace.WorkspaceEditorInterface
	stageView       *editor_stage_view.StageView
	root            *document.Element
	stageViewport   *document.Element
	dimensionToggle *document.Element
	graph           shaderGraph
}

type ShadingWorkspaceUIData struct {
	CameraMode string
}

func (w *ShadingWorkspace) ID() string          { return ID }
func (w *ShadingWorkspace) DisplayName() string { return DisplayName }
func (w *ShadingWorkspace) IsRequired() bool    { return false }

func (w *ShadingWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("ShadingWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	data := ShadingWorkspaceUIData{CameraMode: w.stageView.Camera().ModeString()}
	if err := w.CommonWorkspace.InitializeWithUI(ed.Host(),
		"editor/ui/workspace/shading_workspace.go.html", data, map[string]func(*document.Element){
			"toggleDimension": w.toggleDimension,
		}); err != nil {
		return err
	}
	w.root, _ = w.Doc.GetElementById("shadingWorkspace")
	w.stageViewport, _ = w.Doc.GetElementById("stageViewport")
	w.dimensionToggle, _ = w.Doc.GetElementById("dimensionToggle")
	if w.root != nil {
		w.root.UIPanel.AllowClickThrough()
	}
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

func (w *ShadingWorkspace) Shutdown() {
	defer tracing.NewRegion("ShadingWorkspace.Shutdown").End()
	w.graph.Shutdown()
	w.CommonShutdown()
}

func (w *ShadingWorkspace) Open() {
	defer tracing.NewRegion("ShadingWorkspace.Open").End()
	w.CommonOpen()
	if w.stageViewport != nil {
		w.stageView.SetViewportUI(w.stageViewport.UI)
	}
	if w.dimensionToggle != nil {
		w.dimensionToggle.InnerLabel().SetText(w.stageView.Camera().ModeString())
	}
	w.graph.Open()
	w.stageView.Open()
}

func (w *ShadingWorkspace) Close() {
	defer tracing.NewRegion("ShadingWorkspace.Close").End()
	if w.stageView != nil {
		w.stageView.SetViewportUI(nil)
		w.stageView.Close()
	}
	w.graph.Close()
	w.CommonClose()
}

func (w *ShadingWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *ShadingWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("ShadingWorkspace.Update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred {
		return
	}
	w.graph.Update()
	if w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *ShadingWorkspace) toggleDimension(e *document.Element) {
	defer tracing.NewRegion("ShadingWorkspace.toggleDimension").End()
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
