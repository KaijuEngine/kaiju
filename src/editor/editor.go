package editor

import (
	"kaiju/editor/ui/menu"
	"kaiju/engine"
)

type Editor struct {
	Host *engine.Host
	menu *menu.Menu
}

func New(host *engine.Host) *Editor {
	host.SetFrameRateLimit(60)
	return &Editor{
		Host: host,
	}
}

func (e *Editor) SetupUI() {
	e.menu = menu.New(e.Host)
}
