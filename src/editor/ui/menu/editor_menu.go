package menu

import (
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"os/exec"
	"runtime"
)

type Menu struct {
	doc    *document.Document
	isOpen bool
}

func (m *Menu) close() {
	all := m.doc.GetElementsByClass("menuItemList")
	for i := range all {
		all[i].UI.Entity().Deactivate()
	}
}

func (m *Menu) open(target *document.DocElement) {
	m.close()
	target.UI.Entity().SetActive(m.isOpen)
}

func (m *Menu) openMenu(targetId string) {
	if t, ok := m.doc.GetElementById(targetId); ok {
		m.isOpen = !m.isOpen
		m.open(&t)
	}
}

func (m *Menu) hoverOpenMenu(targetId string) {
	if !m.isOpen {
		return
	}
	if t, ok := m.doc.GetElementById(targetId); ok {
		m.open(&t)
	}
}

func openRepository(*document.DocElement) {
	cmd := "open"
	if runtime.GOOS == "windows" {
		cmd = "explorer"
	}
	exec.Command(cmd, "https://github.com/KaijuEngine/kaiju").Run()
}

func New(host *engine.Host) *Menu {
	html := klib.MustReturn(host.AssetDatabase().ReadText("ui/editor/menu.html"))
	m := &Menu{}
	funcMap := map[string]func(*document.DocElement){
		"openRepository": openRepository,
	}
	m.doc = markup.DocumentFromHTMLString(host, html, "", nil, funcMap)
	allItems := m.doc.GetElementsByClass("menuItem")
	for i := range allItems {
		targetId := allItems[i].HTML.Attribute("data-target")
		allItems[i].UI.AddEvent(ui.EventTypeClick, func() {
			m.openMenu(targetId)
		})
		allItems[i].UI.AddEvent(ui.EventTypeEnter, func() {
			m.hoverOpenMenu(targetId)
		})
		if t, ok := m.doc.GetElementById(targetId); ok {
			l := t.UI.Layout()
			pLayout := allItems[i].UI.Layout()
			allItems[i].UI.AddEvent(ui.EventTypeRender, func() {
				l.SetOffset(pLayout.CalcOffset().X(), l.Offset().Y())
			})
		}
	}
	if b, ok := m.doc.GetElementById("bar"); ok {
		b.UI.AddEvent(ui.EventTypeMiss, func() {
			if m.isOpen {
				m.isOpen = false
				m.close()
			}
		})
	}
	return m
}
