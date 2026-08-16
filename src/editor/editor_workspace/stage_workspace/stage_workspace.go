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
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
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

// init registers the stage workspace singleton with the global registry.
// The editor reads the registry during postProjectLoad to decide which
// workspaces are active.
func init() {
	editor_workspace_registry.Register(&StageWorkspace{})
}

type StageWorkspace struct {
	common_workspace.CommonWorkspace
	ed            editor_workspace.WorkspaceEditorInterface
	stageView     *editor_stage_view.StageView
	pageData      WorkspaceUIData
	hierarchyDoc  *document.Document
	detailsDoc    *document.Document
	contentDoc    *document.Document
	contentUI     WorkspaceContentUI
	hierarchyUI   WorkspaceHierarchyUI
	detailsUI     WorkspaceDetailsUI
	stageViewport stageWorkspaceStageViewport
	cameraPreview stageWorkspaceCameraPreview
	isOpen        bool
	ftde          struct {
		arrow *document.Element
		y     matrix.Float
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
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/stage_workspace.go.html", w.pageData, funcs); err != nil {
		return err
	}
	if err := w.loadPanelDocuments(); err != nil {
		w.destroyPanelDocuments()
		w.CommonShutdown()
		return err
	}
	w.stageViewport.init(&w.UiMan, w.stageView)
	w.cameraPreview.init(&w.UiMan, w.stageView, w)
	w.applyViewportLayout()
	w.hideManualRenderTargetUI()
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

func (w *StageWorkspace) loadPanelDocuments() error {
	defer tracing.NewRegion("StageWorkspace.loadPanelDocuments").End()
	var err error
	w.hierarchyDoc, err = markup.DocumentFromHTMLAsset(&w.UiMan,
		"editor/ui/workspace/stage_workspace_hierarchy.go.html", nil, w.hierarchyUI.setupFuncs())
	if err != nil {
		return err
	}
	w.hierarchyDoc.Deactivate()
	w.detailsDoc, err = markup.DocumentFromHTMLAsset(&w.UiMan,
		"editor/ui/workspace/stage_workspace_details.go.html", nil, w.detailsUI.setupFuncs())
	if err != nil {
		return err
	}
	w.detailsDoc.Deactivate()
	w.contentDoc, err = markup.DocumentFromHTMLAsset(&w.UiMan,
		"editor/ui/workspace/stage_workspace_content.go.html", w.pageData, w.contentUI.setupFuncs())
	if err != nil {
		return err
	}
	w.contentDoc.Deactivate()
	return nil
}

func (w *StageWorkspace) activatePanelDocuments() {
	defer tracing.NewRegion("StageWorkspace.activatePanelDocuments").End()
	for _, doc := range w.panelDocuments() {
		doc.Activate()
	}
}

func (w *StageWorkspace) deactivatePanelDocuments() {
	defer tracing.NewRegion("StageWorkspace.deactivatePanelDocuments").End()
	for _, doc := range w.panelDocuments() {
		doc.Deactivate()
	}
}

func (w *StageWorkspace) markPanelDocumentsDirty() {
	defer tracing.NewRegion("StageWorkspace.markPanelDocumentsDirty").End()
	for _, doc := range w.panelDocuments() {
		doc.MarkDirty()
	}
}

func (w *StageWorkspace) destroyPanelDocuments() {
	defer tracing.NewRegion("StageWorkspace.destroyPanelDocuments").End()
	if w.hierarchyDoc != nil {
		w.hierarchyDoc.Destroy()
		w.hierarchyDoc = nil
	}
	if w.detailsDoc != nil {
		w.detailsDoc.Destroy()
		w.detailsDoc = nil
	}
	if w.contentDoc != nil {
		w.contentDoc.Destroy()
		w.contentDoc = nil
	}
}

func (w *StageWorkspace) panelDocuments() []*document.Document {
	docs := make([]*document.Document, 0, 3)
	if w.hierarchyDoc != nil {
		docs = append(docs, w.hierarchyDoc)
	}
	if w.detailsDoc != nil {
		docs = append(docs, w.detailsDoc)
	}
	if w.contentDoc != nil {
		docs = append(docs, w.contentDoc)
	}
	return docs
}

func (w *StageWorkspace) Shutdown() {
	defer tracing.NewRegion("StageWorkspace.Shutdown").End()
	w.contentUI.contentLoader.Stop()
	if w.ed != nil {
		w.ed.Events().OnRequestOpenStage.Remove(w.openStageSubID)
	}
	w.destroyPanelDocuments()
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
	w.isOpen = true
	w.CommonOpen()
	w.activatePanelDocuments()
	w.contentUI.open()
	w.hierarchyUI.open()
	w.detailsUI.open()
	w.stageViewport.bind(w.stageView)
	w.applyViewportLayout()
	w.stageView.Open()
	w.Doc.MarkDirty()
	w.markPanelDocumentsDirty()
}

func (w *StageWorkspace) Close() {
	defer tracing.NewRegion("StageWorkspace.Close").End()
	w.isOpen = false
	w.contentUI.hideTooltip()
	w.stageView.Close()
	w.hideManualRenderTargetUI()
	w.deactivatePanelDocuments()
	w.CommonClose()
}

func (w *StageWorkspace) Hotkeys() []common_workspace.HotKey {
	return nil
}

func (w *StageWorkspace) FocusRename() bool {
	if len(w.stageView.Manager().Selection()) == 0 {
		return false
	}
	w.detailsUI.focusRename()
	return true
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
	if w.IsBlurred {
		return
	}
	w.stageView.SyncStageViewport()
	if w.UiMan.Group.HasRequests() {
		return
	}
	if !w.IsFocusedOnInput() {
		w.hierarchyUI.updateKeyboardSelection()
	}
	w.detailsUI.update()
	w.stageView.Update(deltaTime, w.ed.Project())
	w.cameraPreview.updatePlacement(w)
}

func (w *StageWorkspace) ToggleViewportSplitFocus() bool {
	defer tracing.NewRegion("StageWorkspace.ToggleViewportSplitFocus").End()
	if w.stageView == nil {
		return false
	}
	w.stageViewport.toggleSplitFocus(w.stageView)
	w.applyViewportLayout()
	w.stageView.SyncStageViewport()
	w.stageView.RefreshTransformGizmoVisibility()
	return true
}

func (w *StageWorkspace) ToggleContentPanel() bool {
	defer tracing.NewRegion("StageWorkspace.ToggleContentPanel").End()
	if w.contentUI.contentArea == nil || w.contentUI.contentArea.UI == nil {
		return false
	}
	if w.contentUI.contentArea.UI.Entity().IsActive() {
		w.contentUI.contentArea.UI.Hide()
		w.hierarchyUI.extendHeight()
		w.detailsUI.extendHeight()
	} else {
		w.contentUI.contentArea.UI.Show()
		w.hierarchyUI.standardHeight()
		w.detailsUI.standardHeight()
	}
	w.applyViewportLayout()
	return true
}

func (w *StageWorkspace) ToggleHierarchyPanel() bool {
	defer tracing.NewRegion("StageWorkspace.ToggleHierarchyPanel").End()
	if w.hierarchyUI.hierarchyArea == nil || w.hierarchyUI.hierarchyArea.UI == nil {
		return false
	}
	if w.hierarchyUI.hierarchyArea.UI.Entity().IsActive() {
		w.hierarchyUI.hierarchyArea.UI.Hide()
	} else {
		w.hierarchyUI.hierarchyArea.UI.Show()
	}
	w.applyViewportLayout()
	return true
}

func (w *StageWorkspace) ToggleDetailsPanel() bool {
	defer tracing.NewRegion("StageWorkspace.ToggleDetailsPanel").End()
	if w.detailsUI.detailsArea == nil || w.detailsUI.detailsArea.UI == nil {
		return false
	}
	if w.detailsUI.detailsArea.UI.Entity().IsActive() {
		w.detailsUI.detailsArea.UI.Hide()
	} else {
		w.detailsUI.detailsArea.UI.Show()
	}
	w.applyViewportLayout()
	return true
}

func (w *StageWorkspace) applyViewportLayout() {
	defer tracing.NewRegion("StageWorkspace.applyViewportLayout").End()
	w.stageViewport.applyLayout(w)
	w.cameraPreview.updatePlacement(w)
}

func (w *StageWorkspace) hideManualRenderTargetUI() {
	w.stageViewport.hide()
	w.cameraPreview.hide()
}

func elementIsActive(elm *document.Element) bool {
	return elm != nil && elm.UI != nil && elm.UI.Entity().IsActive()
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
	w.ftde.y += matrix.Float(deltaTime) * 5
	w.ftde.arrow.UI.Layout().SetOffsetY((1 + matrix.Cos(w.ftde.y)) * 10)
}

func (w *StageWorkspace) removeFtde() {
	defer tracing.NewRegion("StageWorkspace.hideFtde").End()
	if ftde, ok := w.Doc.GetElementById("ftdePrompt"); ok {
		w.Doc.RemoveElement(ftde)
		w.ftde.arrow = nil
	}
}
