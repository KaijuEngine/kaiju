package editor_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/host_container"
)

type EditorWindow interface {
	Tag() string
	Container() *host_container.Container
	Closed()
}

func OpenWindow(win EditorWindow,
	defaultWidth, defaultHeight, defaultX, defaultY int) {
	w, h := defaultWidth, defaultHeight
	x, y := defaultX, defaultY
	if win, err := editor_cache.Window(win.Tag()); err == nil {
		w = win.Width
		h = win.Height
		x = win.X
		y = win.Y
	}
	go win.Container().Run(w, h, x, y)
	<-win.Container().PrepLock
}
