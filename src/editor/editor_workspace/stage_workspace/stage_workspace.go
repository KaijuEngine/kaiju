/******************************************************************************/
/* stage_workspace.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
)

const (
	// ID is the stable workspace identifier used for registration, settings
	// persistence, and SelectWorkspace calls. Exported so other packages can
	// reference the stage workspace without using a magic string.
	ID = "stage"

	// DisplayName is the label shown on the stage workspace's menu bar tab.
	DisplayName = "Stage"

	maxContentDropDistance = 10
)

type stageViewportLayoutMode int

const (
	stageViewportLayoutSingle stageViewportLayoutMode = iota
	stageViewportLayoutQuad
)

// init registers the stage workspace singleton with the global registry.
// The editor reads the registry during postProjectLoad to decide which
// workspaces are active.
func init() {
	editor_workspace_registry.Register(&StageWorkspace{})
}

type StageWorkspace struct {
	common_workspace.CommonWorkspace
	ed          editor_workspace.WorkspaceEditorInterface
	stageView   *editor_stage_view.StageView
	pageData    WorkspaceUIData
	contentUI   WorkspaceContentUI
	hierarchyUI WorkspaceHierarchyUI
	detailsUI   WorkspaceDetailsUI
	viewports   map[editor_stage_view.StageViewportKind]*document.Element
	labels      map[editor_stage_view.StageViewportKind]*document.Element
	layoutMode  stageViewportLayoutMode
	focusedView editor_stage_view.StageViewportKind
	ftde        struct {
		arrow *document.Element
		y     float32
	}
	openStageSubID events.Id
}

func (w *StageWorkspace) ID() string          { return ID }
func (w *StageWorkspace) DisplayName() string { return DisplayName }
func (w *StageWorkspace) IsRequired() bool    { return true }

func (w *StageWorkspace) Initialize(ed editor_workspace.WorkspaceEditorInterface) error {
	defer tracing.NewRegion("StageWorkspace.Initialize").End()
	host := ed.Host()
	w.ed = ed
	w.stageView = ed.StageView()
	w.stageView.Initialize(host, ed)
	w.pageData.SetupUIData(w.ed.Cache(), ed.StageView().Camera().ModeString())
	funcs := map[string]func(*document.Element){
		"toggleDimension": w.toggleDimension,
	}
	funcs = klib.MapJoin(funcs, w.contentUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.hierarchyUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.detailsUI.setupFuncs())
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/stage_workspace.go.html", w.pageData, funcs); err != nil {
		return err
	}
	viewportIDs := []struct {
		kind editor_stage_view.StageViewportKind
		id   string
	}{
		{editor_stage_view.StageViewportPerspective, "stageViewport"},
		{editor_stage_view.StageViewportTop, "stageViewportTop"},
		{editor_stage_view.StageViewportSide, "stageViewportSide"},
		{editor_stage_view.StageViewportFront, "stageViewportFront"},
	}
	w.viewports = make(map[editor_stage_view.StageViewportKind]*document.Element, len(viewportIDs))
	for _, viewportID := range viewportIDs {
		if viewport, ok := w.Doc.GetElementById(viewportID.id); ok {
			w.viewports[viewportID.kind] = viewport
			w.stageView.SetViewportUIForKind(viewportID.kind, viewport.UI)
		}
	}
	labelIDs := []struct {
		kind editor_stage_view.StageViewportKind
		id   string
	}{
		{editor_stage_view.StageViewportPerspective, "stageViewportLabelPerspective"},
		{editor_stage_view.StageViewportTop, "stageViewportLabelTop"},
		{editor_stage_view.StageViewportSide, "stageViewportLabelSide"},
		{editor_stage_view.StageViewportFront, "stageViewportLabelFront"},
	}
	w.labels = make(map[editor_stage_view.StageViewportKind]*document.Element, len(labelIDs))
	for _, labelID := range labelIDs {
		if label, ok := w.Doc.GetElementById(labelID.id); ok {
			w.labels[labelID.kind] = label
		}
	}
	w.layoutMode = stageViewportLayoutSingle
	w.focusedView = editor_stage_view.StageViewportPerspective
	w.applyViewportLayout()
	w.ftde.arrow, _ = w.Doc.GetElementById("ftdeArrow")
	w.contentUI.setup(w, w.ed.Events())
	w.hierarchyUI.setup(w)
	w.detailsUI.setup(w)
	// Subscribe to cross-workspace requests. The content workspace publishes
	// OnRequestOpenStage when the user picks a stage asset; we open it and
	// switch ourselves active.
	w.openStageSubID = ed.Events().OnRequestOpenStage.Add(func(stageID string) {
		w.OpenStage(stageID)
		ed.SelectWorkspace(ID)
	})
	w.loadLastOpenStage()
	return nil
}

func (w *StageWorkspace) Shutdown() {
	defer tracing.NewRegion("StageWorkspace.Shutdown").End()
	if w.ed != nil {
		w.ed.Events().OnRequestOpenStage.Remove(w.openStageSubID)
	}
	w.CommonShutdown()
}

func (w *StageWorkspace) loadLastOpenStage() {
	defer tracing.NewRegion("StageWorkspace.loadLastOpenStage").End()
	p := w.ed.Project()
	lastStage := p.Settings.EditorSettings.LatestOpenStage
	if lastStage != "" {
		w.OpenStage(lastStage)
	}
}

func (w *StageWorkspace) Open() {
	defer tracing.NewRegion("StageWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.contentUI.open()
	w.hierarchyUI.open()
	w.detailsUI.open()
	w.Doc.MarkDirty()
}

func (w *StageWorkspace) Close() {
	defer tracing.NewRegion("StageWorkspace.Close").End()
	w.stageView.Close()
	w.CommonClose()
}

func (w *StageWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{
		{
			Keys: []hid.KeyboardKey{hid.KeyboardKeyF2},
			Call: w.focusRename,
		},
	}
}

func (w *StageWorkspace) focusRename() {
	if len(w.stageView.Manager().Selection()) == 0 {
		return
	}
	w.detailsUI.focusRename()
}

func (w *StageWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("StageWorkspace.update").End()
	w.updateFtde(deltaTime)
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if windowing.HasDragData() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.detailsUI.update()
	didViewportToggle := w.processViewportLayoutHotkeys()
	didKeyboardActions := w.stageView.Update(deltaTime, w.ed.Project())
	if !didViewportToggle && !didKeyboardActions {
		w.contentUI.processHotkeys(w.Host)
		w.hierarchyUI.processHotkeys(w.Host)
		w.detailsUI.processHotkeys(w.Host)
	}
}

func (w *StageWorkspace) processViewportLayoutHotkeys() bool {
	defer tracing.NewRegion("StageWorkspace.processViewportLayoutHotkeys").End()
	kb := &w.Host.Window.Keyboard
	if !kb.KeyDown(hid.KeyboardKeyP) || kb.HasModifier() {
		return false
	}
	if w.layoutMode == stageViewportLayoutSingle {
		w.layoutMode = stageViewportLayoutQuad
	} else {
		focused, ok := w.stageView.HoveredViewportKind()
		if !ok {
			focused, ok = w.stageView.ActiveViewportKind()
		}
		if !ok {
			focused = editor_stage_view.StageViewportPerspective
		}
		w.focusedView = focused
		w.stageView.FocusViewportKind(focused)
		w.layoutMode = stageViewportLayoutSingle
	}
	w.applyViewportLayout()
	return true
}

func (w *StageWorkspace) applyViewportLayout() {
	defer tracing.NewRegion("StageWorkspace.applyViewportLayout").End()
	for _, kind := range editor_stage_view.StageViewportKinds() {
		if viewport := w.viewports[kind]; viewport != nil {
			w.Doc.SetElementClassesWithoutApply(viewport, w.viewportClasses(kind)...)
		}
		if label := w.labels[kind]; label != nil {
			w.Doc.SetElementClassesWithoutApply(label, w.viewportLabelClasses(kind)...)
		}
	}
	for _, kind := range editor_stage_view.StageViewportKinds() {
		visible := w.viewportVisible(kind)
		if viewport := w.viewports[kind]; viewport != nil {
			viewport.UI.SetVisibility(visible)
		}
		if label := w.labels[kind]; label != nil {
			label.UI.SetVisibility(visible)
		}
	}
	w.Doc.ApplyStyles()
}

func (w *StageWorkspace) viewportVisible(kind editor_stage_view.StageViewportKind) bool {
	return w.layoutMode == stageViewportLayoutQuad || kind == w.focusedView
}

func (w *StageWorkspace) viewportClasses(kind editor_stage_view.StageViewportKind) []string {
	if w.layoutMode == stageViewportLayoutSingle {
		if kind == w.focusedView {
			return []string{"stageViewport", "stageViewportSingle"}
		}
		return []string{"stageViewport", "stageViewportHidden"}
	}
	return []string{"stageViewport", viewportQuadClass(kind)}
}

func (w *StageWorkspace) viewportLabelClasses(kind editor_stage_view.StageViewportKind) []string {
	if w.layoutMode == stageViewportLayoutSingle {
		if kind == w.focusedView {
			return []string{"stageViewportLabel", "stageViewportLabelSingle"}
		}
		return []string{"stageViewportLabel", "stageViewportLabelHidden"}
	}
	return []string{"stageViewportLabel", viewportQuadLabelClass(kind)}
}

func viewportQuadClass(kind editor_stage_view.StageViewportKind) string {
	switch kind {
	case editor_stage_view.StageViewportTop:
		return "stageViewportQuadTop"
	case editor_stage_view.StageViewportSide:
		return "stageViewportQuadSide"
	case editor_stage_view.StageViewportFront:
		return "stageViewportQuadFront"
	default:
		return "stageViewportQuadPerspective"
	}
}

func viewportQuadLabelClass(kind editor_stage_view.StageViewportKind) string {
	switch kind {
	case editor_stage_view.StageViewportTop:
		return "stageViewportLabelQuadTop"
	case editor_stage_view.StageViewportSide:
		return "stageViewportLabelQuadSide"
	case editor_stage_view.StageViewportFront:
		return "stageViewportLabelQuadFront"
	default:
		return "stageViewportLabelQuadPerspective"
	}
}

func (w *StageWorkspace) toggleDimension(e *document.Element) {
	defer tracing.NewRegion("StageWorkspace.toggleDimension").End()
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

func (w *StageWorkspace) updateFtde(deltaTime float64) {
	defer tracing.NewRegion("StageWorkspace.updateFtde").End()
	if w.ftde.arrow == nil {
		return
	}
	w.ftde.y += float32(deltaTime) * 5
	w.ftde.arrow.UI.Layout().SetOffsetY((1 + matrix.Cos(w.ftde.y)) * 10)
}

func (w *StageWorkspace) removeFtde() {
	defer tracing.NewRegion("StageWorkspace.hideFtde").End()
	if ftde, ok := w.Doc.GetElementById("ftdePrompt"); ok {
		w.Doc.RemoveElement(ftde)
		w.ftde.arrow = nil
	}
}
