/******************************************************************************/
/* stage_viewport_interaction.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

func (v *StageView) processViewportInteractions(proj *project.Project) {
	defer tracing.NewRegion("StageWorkspace.processViewportInteractions").End()
	m := &v.host.Window.Mouse
	kb := &v.host.Window.Keyboard
	insideViewport := v.viewportContainsScreenPosition(m.ScreenPosition())
	if !insideViewport && !v.selectTool.IsActive() && !v.transformTool.IsActive() &&
		!v.transformMan.IsBusy() && !v.vertexSnap.IsBusy() {
		return
	}
	if v.toolOwner != nil && v.toolOwner.UpdateViewportTool(v) {
		return
	}
	if v.vertexSnap.Update(v.host) {
		return
	}
	if v.transformTool.Update() {
		return
	}
	v.transformMan.Update(v.host, proj)
	if v.transformMan.IsBusy() {
		return
	}
	boxSelected := v.selectTool.Update()
	if m.Released(hid.MouseButtonLeft) && !boxSelected {
		ray := v.activeCamera().RayCast(m)
		mode := stageSelectionMode(kb)
		point := v.ViewportMousePosition(m)
		if v.stagePicking.RequestClick(point, mode, ray) {
			return
		}
		if mode == editor_stage_manager.SelectionModeAppend {
			v.manager.TryAppendSelect(ray)
		} else if mode == editor_stage_manager.SelectionModeToggle {
			v.manager.TryToggleSelect(ray)
		} else {
			v.manager.TrySelect(ray)
		}
	}
	if kb.KeyDown(hid.KeyboardKeyF) {
		if v.manager.HasSelection() {
			v.activeCamera().Focus(v.manager.SelectionBounds())
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
	return v.activeCamera().Mode() == editor_controls.EditorCameraMode3d
}
