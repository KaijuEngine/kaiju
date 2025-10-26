package editor

import (
	"time"
	"weak"
)

func (ed *Editor) setupWindowActivity() {
	wed := weak.Make(ed)
	ed.window.activateId = ed.host.Window.OnActivate.Add(func() {
		sed := wed.Value()
		if sed != nil {
			sed.onWindowActivate()
		}
	})
	ed.window.deactivateId = ed.host.Window.OnDeactivate.Add(func() {
		sed := wed.Value()
		if sed != nil {
			sed.onWindowDeactivate()
		}
	})
}

func (ed *Editor) onWindowActivate() {
	if !time.Now().After(ed.window.lastActiveTime.Add(time.Second * 5)) {
		return
	}
	go ed.project.ReadSourceCode()
}

func (ed *Editor) onWindowDeactivate() {
	ed.window.lastActiveTime = time.Now()
}
