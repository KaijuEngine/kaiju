package stage_workspace

import (
	"kaiju/editor/editor_workspace/stage_workspace/transform_tools"
	"kaiju/platform/hid"
)

func (w *Workspace) processViewportInteractions() {
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
