package menu

import (
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
)

type Menu struct {
	doc *document.Document
}

func (m *Menu) openMenu(elm *document.DocElement) {
	elm.HTML.Children[0].DocumentElement.UI.Entity().Activate()
}

func New(host *engine.Host) *Menu {
	html := klib.MustReturn(host.AssetDatabase().ReadText("ui/editor/menu.html"))
	m := &Menu{}
	funcMap := map[string]func(*document.DocElement){
		"openMenu": m.openMenu,
	}
	m.doc = markup.DocumentFromHTMLString(host, html, "", nil, funcMap)
	return m
}
