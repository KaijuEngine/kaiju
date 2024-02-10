package editor

import (
	"kaiju/editor/ui/filesystem_select"
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

func (e *Editor) testBtn(*document.DocElement) {
	filesystem_select.New(func(s string) {
		println(s)
	})
	//search := windowing.FileSearch{
	//	Title:     "Binary file",
	//	Extension: "bin",
	//}
	//if s, ok := e.Host.Window.OpenFile(search); ok {
	//	println(s)
	//} else {
	//	println("no file selected")
	//}
}

func (e *Editor) SetupUI() {
	e.Host.CreatingEditorEntities()
	//e.menu = menu.New(e.Host)
	html := klib.MustReturn(e.Host.AssetDatabase().ReadText("ui/editor/project.html"))
	markup.DocumentFromHTMLString(e.Host, html, "", nil,
		map[string]func(*document.DocElement){
			"testBtn": e.testBtn,
		})
	e.Host.DoneCreatingEditorEntities()
}
