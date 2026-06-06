/******************************************************************************/
/* editor_window.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"time"
	"weak"

	"kaijuengine.com/platform/profiler/tracing"
)

func (ed *Editor) setupWindowActivity() {
	defer tracing.NewRegion("Editor.setupWindowActivity").End()
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
	// goroutine
	go ed.project.ReadSourceCode()
}

func (ed *Editor) onWindowDeactivate() {
	ed.window.lastActiveTime = time.Now()
}
