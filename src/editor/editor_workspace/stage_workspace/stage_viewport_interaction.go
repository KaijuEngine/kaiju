/******************************************************************************/
/* stage_viewport_interaction.go                                              */
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
	"kaiju/editor/editor_workspace/stage_workspace/transform_tools"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
)

func (w *Workspace) processViewportInteractions() {
	defer tracing.NewRegion("StageWorkspace.processViewportInteractions").End()
	m := &w.Host.Window.Mouse
	kb := &w.Host.Window.Keyboard
	if m.Pressed(hid.MouseButtonLeft) {
		ray := w.camera.RayCast(m)
		if kb.HasShift() {
			w.manager.TryAppendSelect(ray)
		} else if kb.HasCtrl() {
			w.manager.TryToggleSelect(ray)
		} else {
			w.manager.TrySelect(ray)
		}
	}
	if w.transformTool.Update() {
		return
	}
	if kb.KeyDown(hid.KeyboardKeyF) {
		if w.manager.HasSelection() {
			w.camera.Focus(w.manager.SelectionBounds())
		}
	} else if kb.KeyDown(hid.KeyboardKeyG) {
		w.transformTool.Enable(transform_tools.ToolStateMove)
	} else if kb.KeyDown(hid.KeyboardKeyR) {
		w.transformTool.Enable(transform_tools.ToolStateRotate)
	} else if kb.KeyDown(hid.KeyboardKeyS) {
		w.transformTool.Enable(transform_tools.ToolStateScale)
	}
}
