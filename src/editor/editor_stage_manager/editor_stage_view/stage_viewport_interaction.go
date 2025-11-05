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

package editor_stage_view

import (
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
)

func (v *StageView) processViewportInteractions() {
	defer tracing.NewRegion("StageWorkspace.processViewportInteractions").End()
	m := &v.host.Window.Mouse
	kb := &v.host.Window.Keyboard
	if v.transformTool.Update() {
		return
	}
	if m.Pressed(hid.MouseButtonLeft) {
		ray := v.camera.RayCast(m)
		if kb.HasShift() {
			v.manager.TryAppendSelect(ray)
		} else if kb.HasCtrl() {
			v.manager.TryToggleSelect(ray)
		} else {
			v.manager.TrySelect(ray)
		}
	}
	if kb.KeyDown(hid.KeyboardKeyF) {
		if v.manager.HasSelection() {
			v.camera.Focus(v.manager.SelectionBounds())
		}
	} else if kb.KeyDown(hid.KeyboardKeyG) {
		v.transformTool.Enable(transform_tools.ToolStateMove)
	} else if kb.KeyDown(hid.KeyboardKeyR) {
		v.transformTool.Enable(transform_tools.ToolStateRotate)
	} else if kb.KeyDown(hid.KeyboardKeyS) {
		v.transformTool.Enable(transform_tools.ToolStateScale)
	}
}

func (v *StageView) isCamera3D() bool {
	return v.camera.Mode() == editor_controls.EditorCameraMode3d
}
