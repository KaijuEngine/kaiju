package editor_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/klib"
)

type Listing struct {
	windows []EditorWindow
}

func New() Listing {
	return Listing{
		windows: make([]EditorWindow, 0),
	}
}

func (l *Listing) Add(w EditorWindow) {
	l.windows = append(l.windows, w)
	w.Container().Host.OnClose.Add(func() {
		saveLayout(w, false)
		w.Closed()
		l.Remove(w)
	})
}

func (l *Listing) Remove(w EditorWindow) {
	for i, win := range l.windows {
		if win == w {
			l.windows = klib.RemoveUnordered(l.windows, i)
			break
		}
	}
}

func (l *Listing) CloseAll() {
	for _, win := range l.windows {
		saveLayout(win, true)
		win.Container().Host.Close()
	}
	l.windows = l.windows[:0]
}

func saveLayout(win EditorWindow, isOpen bool) {
	host := win.Container().Host
	x := host.Window.X()
	y := host.Window.Y()
	w := host.Window.Width()
	h := host.Window.Height()
	editor_cache.SetWindow(win.Tag(), x, y, w, h, isOpen)
}
