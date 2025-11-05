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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"kaiju/platform/profiler/tracing"
)

const maxContentDropDistance = 10

type Workspace struct {
	common_workspace.CommonWorkspace
	ed          StageWorkspaceEditorInterface
	stageView   *editor_stage_view.StageView
	pageData    content_workspace.WorkspaceUIData
	contentUI   WorkspaceContentUI
	hierarchyUI WorkspaceHierarchyUI
	detailsUI   WorkspaceDetailsUI
	updateId    engine.UpdateId
}

func (w *Workspace) Initialize(host *engine.Host, ed StageWorkspaceEditorInterface) {
	defer tracing.NewRegion("StageWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	w.stageView.Initialize(host, ed.History(), &ed.Settings().Snapping)
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

}

func (w *Workspace) Open() {
	defer tracing.NewRegion("StageWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.updateId = w.Host.Updater.AddUpdate(w.update)
	w.contentUI.open()
	w.hierarchyUI.open()
	w.detailsUI.open()
	w.Host.RunOnMainThread(w.Doc.Clean)
}

func (w *Workspace) Close() {
	defer tracing.NewRegion("StageWorkspace.Close").End()
	w.stageView.Close()
	w.Host.Updater.RemoveUpdate(&w.updateId)
	w.CommonClose()
}

func (w *Workspace) update(deltaTime float64) {
	defer tracing.NewRegion("StageWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if !w.contentUI.update(w) {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.contentUI.processHotkeys(w.Host)
	w.hierarchyUI.processHotkeys(w.Host)
	w.detailsUI.processHotkeys(w.Host)
	w.stageView.Update(deltaTime)
}

func (w *Workspace) toggleDimension(e *document.Element) {
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
