/******************************************************************************/
/* stage_workspace.go                                                         */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
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
		"toggleDimension":           w.toggleDimension,
		"setOcclusionMode":          w.setOcclusionMode,
		"setOcclusionVisualization": w.setOcclusionVisualization,
	}
	funcs = klib.MapJoin(funcs, w.contentUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.hierarchyUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.detailsUI.setupFuncs())
	if err := w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/stage_workspace.go.html", w.pageData, funcs); err != nil {
		return err
	}
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
	w.updateOcclusionDebugUI()
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
	w.updateOcclusionDebugUI()
	w.detailsUI.update()
	didKeyboardActions := w.stageView.Update(deltaTime, w.ed.Project())
	if !didKeyboardActions {
		w.contentUI.processHotkeys(w.Host)
		w.hierarchyUI.processHotkeys(w.Host)
		w.detailsUI.processHotkeys(w.Host)
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
