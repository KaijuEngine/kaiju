/******************************************************************************/
/* vfx_workspace.go                                                           */
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

package vfx_workspace

import (
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/vfx"
)

type VfxWorkspace struct {
	common_workspace.CommonWorkspace
	ed        VfxWorkspaceEditorInterface
	stageView *editor_stage_view.StageView
	updateId  engine.UpdateId
}

func (w *VfxWorkspace) Initialize(host *engine.Host, ed VfxWorkspaceEditorInterface) {
	defer tracing.NewRegion("VfxWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/vfx_workspace.go.html", nil, map[string]func(*document.Element){
			"clickTest": w.clickTest,
		})
}

func (w *VfxWorkspace) Open() {
	defer tracing.NewRegion("VfxWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.updateId = w.Host.Updater.AddUpdate(w.update)
}

func (w *VfxWorkspace) Close() {
	defer tracing.NewRegion("VfxWorkspace.Close").End()
	w.CommonClose()
	w.stageView.Close()
	w.Host.Updater.RemoveUpdate(&w.updateId)
}

func (w *VfxWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *VfxWorkspace) update(deltaTime float64) {
	defer tracing.NewRegion("VfxWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *VfxWorkspace) clickTest(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.clickTest").End()
	em := &vfx.Emitter{}
	tex, _ := w.Host.TextureCache().Texture("smoke.png", rendering.TextureFilterLinear)
	em.Initialize(w.Host, tex, vfx.EmitterConfig{
		SpawnRate:        0.05,
		ParticleLifeSpan: 2,
		DirectionMin:     matrix.NewVec3(-0.3, 1, -0.3),
		DirectionMax:     matrix.NewVec3(0.3, 1, 0.3),
		VelocityMinMax:   matrix.Vec2One().Scale(1),
		OpacityMinMax:    matrix.NewVec2(0.3, 1.0),
		FadeOutOverLife:  true,
	})
}
