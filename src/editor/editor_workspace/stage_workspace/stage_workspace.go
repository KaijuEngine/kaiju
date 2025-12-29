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
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/editor_workspace/content_workspace"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/platform/windowing"
)

const maxContentDropDistance = 10

type StageWorkspace struct {
	common_workspace.CommonWorkspace
	ed          StageWorkspaceEditorInterface
	stageView   *editor_stage_view.StageView
	pageData    content_workspace.WorkspaceUIData
	contentUI   WorkspaceContentUI
	hierarchyUI WorkspaceHierarchyUI
	detailsUI   WorkspaceDetailsUI
}

func (w *StageWorkspace) Initialize(host *engine.Host, ed StageWorkspaceEditorInterface) {
	defer tracing.NewRegion("StageWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	w.stageView.Initialize(host, ed.History(), ed.Settings(), ed)
	w.pageData.SetupUIData(w.ed.Cache())
	funcs := map[string]func(*document.Element){
		"toggleDimension": w.toggleDimension,
	}
	funcs = klib.MapJoin(funcs, w.contentUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.hierarchyUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.detailsUI.setupFuncs())
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/stage_workspace.go.html", w.pageData, funcs)
	w.contentUI.setup(w, w.ed.Events())
	w.hierarchyUI.setup(w)
	w.detailsUI.setup(w)
	w.initLLMActions()
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
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if windowing.HasDragData() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
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
