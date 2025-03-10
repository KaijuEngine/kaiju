package tab_container

import (
	"kaiju/engine"
	"kaiju/engine/globals"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"slices"
	"weak"
)

type TabContainerTab struct {
	Label  string
	parent weak.Pointer[TabContainer]
}

type TabContainer struct {
	host  *engine.Host
	doc   *document.Document
	Tabs  []TabContainerTab
	uiMan ui.Manager
}

func (t *TabContainerTab) DragUpdate() {}

func (t *TabContainer) findTab(label string) *TabContainerTab {
	for i := range t.Tabs {
		if t.Tabs[i].Label == label {
			return &t.Tabs[i]
		}
	}
	return nil
}

func (t *TabContainer) removeTab(tab *TabContainerTab) {
	for i := range t.Tabs {
		if &t.Tabs[i] == tab {
			t.Tabs = slices.Delete(t.Tabs, i, i+1)
		}
	}
}

func (t *TabContainer) tabDragStart(e *document.Element) {
	win := t.host.Window
	globals.SetDragData(t.findTab(e.Attribute("id")))
	win.CursorSizeAll()
}

func (t *TabContainer) tabDragEnter(e *document.Element) {
	tex, err := t.host.TextureCache().Texture("textures/window_tab_drag_enter.png",
		rendering.TextureFilterNearest)
	if err == nil {
		e.UI.ToPanel().SetBackground(tex)
	}
}

func (t *TabContainer) tabDragLeave(e *document.Element) {
	tex, err := t.host.TextureCache().Texture("textures/window_tab.png",
		rendering.TextureFilterNearest)
	if err == nil {
		e.UI.ToPanel().SetBackground(tex)
	}
}

func (t *TabContainer) tabDrop(e *document.Element) {
	dd := globals.DragData()
	tab, ok := dd.(*TabContainerTab)
	if !ok {
		return
	}
	lastParent := tab.parent.Value()
	onTab := t.findTab(e.Attribute("id"))
	if tab == onTab {
		return
	}
	if lastParent != t {
		for i := range t.Tabs {
			if &t.Tabs[i] == onTab {
				t.Tabs = slices.Insert(t.Tabs, i, *tab)
				break
			}
		}
		tab.parent = weak.Make(t)
		lastParent.removeTab(tab)
	} else {
		from := 0
		to := 0
		for i := range t.Tabs {
			if &t.Tabs[i] == tab {
				from = i
			} else if &t.Tabs[i] == onTab {
				to = i
			}
		}
		klib.SliceMove(t.Tabs, from, to)
		t.reload()
	}
}

func (t *TabContainer) tabDragEnterRoot(*document.Element) {}
func (t *TabContainer) tabDragLeaveRoot(*document.Element) {}
func (t *TabContainer) tabDropRoot(*document.Element)      {}

func (t *TabContainer) tabClick(*document.Element) {}

func (t *TabContainer) reload() {
	const html = "editor/ui/tab_container/tab_container.html"
	if t.doc != nil {
		t.doc.Destroy()
	}
	t.doc, _ = markup.DocumentFromHTMLAsset(&t.uiMan, html, t, map[string]func(*document.Element){
		"tabClick":         t.tabClick,
		"tabDragStart":     t.tabDragStart,
		"tabDragEnterRoot": t.tabDragEnterRoot,
		"tabDragLeaveRoot": t.tabDragLeaveRoot,
		"tabDragEnter":     t.tabDragEnter,
		"tabDragLeave":     t.tabDragLeave,
		"tabDrop":          t.tabDrop,
		"tabDropRoot":      t.tabDropRoot,
	})
}

func New(tests []string) {
	container := host_container.New("Tab Container", nil)
	go container.Run(500, 300, -1, -1)
	<-container.PrepLock
	t := &TabContainer{
		host: container.Host,
	}
	t.uiMan.Init(container.Host)
	for i := range tests {
		t.Tabs = append(t.Tabs, TabContainerTab{tests[i], weak.Make(t)})
	}
	container.RunFunction(func() { t.reload() })
}
