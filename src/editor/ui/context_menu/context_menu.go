package context_menu

import (
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
)

type ContextMenu struct {
	container *host_container.Container
	doc       *document.Document
	uiMan     *ui.Manager
	entries   []ContextMenuEntry
	x         float32
	y         float32
}

type ContextMenuEntry struct {
	Id      string
	Label   string
	OnClick func()
}

func New(container *host_container.Container, uiMan *ui.Manager) *ContextMenu {
	c := &ContextMenu{
		container: container,
		uiMan:     uiMan,
		entries:   []ContextMenuEntry{},
	}
	return c
}

func (c *ContextMenu) reload() {
	c.Hide()
	c.container.Host.CreatingEditorEntities()
	html := klib.MustReturn(c.container.Host.AssetDatabase().ReadText("editor/ui/context_menu.html"))
	funcMap := map[string]func(*document.Element){
		"selectEntry": c.selectEntry,
		"clickMiss":   c.clickMiss,
	}
	c.doc = markup.DocumentFromHTMLString(c.uiMan, html, "", c.entries, funcMap)
	m, _ := c.doc.GetElementById("contextMenu")
	c.container.Host.DoneCreatingEditorEntities()
	ww := float32(c.container.Host.Window.Width())
	wh := float32(c.container.Host.Window.Height())
	ps := m.UIPanel.Base().Layout().PixelSize()
	if c.x+ps.Width() > ww {
		c.x = ww - ps.Width()
	}
	if c.y+ps.Height() > wh {
		c.y = wh - ps.Height()
	}
	m.UIPanel.Base().Layout().SetOffset(c.x, c.y)
}

func (c *ContextMenu) Show(entries []ContextMenuEntry) {
	mouse := &c.container.Host.Window.Mouse
	c.entries = entries
	c.x = mouse.ScreenPosition().X()
	c.y = mouse.ScreenPosition().Y()
	c.reload()
}

func (c *ContextMenu) Hide() {
	if c.doc != nil {
		c.doc.Destroy()
	}
	c.doc = nil
}

func (c *ContextMenu) clickMiss(*document.Element) { c.Hide() }

func (c *ContextMenu) selectEntry(elm *document.Element) {
	id := elm.Attribute("id")
	for i := range c.entries {
		if c.entries[i].Id == id {
			c.entries[i].OnClick()
		}
	}
	c.Hide()
}
