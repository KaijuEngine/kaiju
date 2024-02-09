package editor

import (
	"kaiju/editor/ui/menu"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
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
	e.Host.CreatingEditorEntities()
	//e.menu = menu.New(e.Host)
	html := klib.MustReturn(e.Host.AssetDatabase().ReadText("ui/editor/project.html"))
	markup.DocumentFromHTMLString(e.Host, html, "", nil, map[string]func(*document.DocElement){
		"testBtn": func(*document.DocElement) {
			if s, ok := e.Host.Window.OpenFile(".bin"); ok {
				println(s)
			} else {
				println("no file selected")
			}
		},
	})
	e.Host.DoneCreatingEditorEntities()
}
