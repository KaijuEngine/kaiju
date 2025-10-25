package stage_workspace

import (
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
	if kb.KeyDown(hid.KeyboardKeyF) {
		if w.manager.HasSelection() {
			w.camera.Focus(w.manager.SelectionBounds())

		}
	}
}
