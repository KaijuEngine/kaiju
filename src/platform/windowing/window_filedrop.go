//go:build editor || filedrop

/******************************************************************************/
/* window_filedrop.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

/******************************************************************************/
/* window_filedrop.go                                                         */
/******************************************************************************/

package windowing

import (
	"log/slog"
	"slices"
	"sync"

	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/platform/profiler/tracing"
)

// FileDropEvent keeps a native drop as one batch so higher layers can route it
// by location before deciding how the files should be processed.
//
// NOTE: position (x, y) is relative to client area
type FileDropEvent struct {
	X     int
	Y     int
	Paths []string
}

type fileDropModule struct {
	onDrop  events.EventWithArg[FileDropEvent]
	pending []FileDropEvent
	mutex   sync.Mutex
}

func (w *Window) OnFileDrop() *events.EventWithArg[FileDropEvent] {
	return &w.fileDrop.onDrop
}

func (w *Window) SetFileDropEnabled(enabled bool) {
	w.setFileDropEnabled(enabled)
}

func (m *fileDropModule) addFileDropToQueue(evt FileDropEvent) {
	defer tracing.NewRegion("fileDropModule.addFileDropToQueue").End()
	evt.Paths = slices.Clone(evt.Paths)
	m.mutex.Lock()
	m.pending = append(m.pending, evt)
	m.mutex.Unlock()
}

func (m *fileDropModule) processQueuedFileDrops() {
	defer tracing.NewRegion("fileDropModule.processQueuedFileDrops").End()
	m.mutex.Lock()
	pending := slices.Clone(m.pending)
	m.pending = m.pending[:0]
	m.mutex.Unlock()
	for i := range pending {
		m.onDrop.Execute(pending[i])
	}
}

// NOTE: position (x, y) is relative to client area
func queueNativeFileDropEvent(goWindow uint64, x, y int, paths []string) {
	defer tracing.NewRegion("windowing.queueNativeFileDropEvent").End()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic while enqueueing file drop", "panic", r)
		}
	}()
	if len(paths) == 0 {
		return
	}
	gw, ok := windowLookup.Load(goWindow)
	if !ok || gw == nil {
		return
	}
	win := gw.(*Window)
	win.fileDrop.addFileDropToQueue(FileDropEvent{
		X:     x,
		Y:     y,
		Paths: paths,
	})
}
