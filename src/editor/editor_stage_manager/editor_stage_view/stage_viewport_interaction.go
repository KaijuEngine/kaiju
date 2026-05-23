/******************************************************************************/
/* stage_viewport_interaction.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

const menuBarHeightArea = 30

func (v *StageView) processViewportInteractions(proj *project.Project) {
	defer tracing.NewRegion("StageWorkspace.processViewportInteractions").End()
	m := &v.host.Window.Mouse
	kb := &v.host.Window.Keyboard
	// TODO:  This is to prevent deselecting and box selection if the mouse was
	// over the menu bar area. Probably should do this check in a better way
	if m.ScreenPosition().Y() <= menuBarHeightArea {
		return
	}
	if v.toolOwner != nil && v.toolOwner.UpdateViewportTool(v) {
		return
	}
	if v.transformTool.Update() {
		return
	}
	v.transformMan.Update(v.host, proj)
	if v.transformMan.IsBusy() {
		return
	}
	v.selectTool.Update()
	if m.Pressed(hid.MouseButtonLeft) {
		ray := v.camera.RayCast(m)
		if kb.HasShift() {
			v.manager.TryAppendSelect(ray)
		} else if kb.HasCtrlOrMeta() {
			v.manager.TryToggleSelect(ray)
		} else {
			v.manager.TrySelect(ray)
		}
	}
	if kb.KeyDown(hid.KeyboardKeyF) {
		if v.manager.HasSelection() {
			v.camera.Focus(v.manager.SelectionBounds())
		}
	} else if kb.KeyDown(hid.KeyboardKey1) {
		v.transformTool.Enable(transform_tools.ToolStateMove)
	} else if kb.KeyDown(hid.KeyboardKey2) {
		v.transformTool.Enable(transform_tools.ToolStateRotate)
	} else if kb.KeyDown(hid.KeyboardKey3) {
		v.transformTool.Enable(transform_tools.ToolStateScale)
	}
}

func (v *StageView) isCamera3D() bool {
	return v.camera.Mode() == editor_controls.EditorCameraMode3d
}
