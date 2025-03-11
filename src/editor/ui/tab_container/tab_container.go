package tab_container

import (
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/ui"
	"kaiju/windowing"
	"slices"
	"weak"
)

type Snap = string

const (
	SnapCenter = Snap("center")
	SnapLeft   = Snap("left")
	SnapTop    = Snap("top")
	SnapRight  = Snap("right")
	SnapBottom = Snap("bottom")
)

type TabContainer struct {
	host      *engine.Host
	doc       *document.Document
	activeTab int
	uiMan     *ui.Manager
	nextTabId int
	Tabs      []TabContainerTab
	Snap      string
}

func (t *TabContainer) tabPtrIndex(tab *TabContainerTab) int {
	for i := range t.Tabs {
		if &t.Tabs[i] == tab {
			return i
		}
	}
	return -1
}

func (t *TabContainer) tabIndex(label string) int {
	for i := range t.Tabs {
		if t.Tabs[i].Label == label {
			return i
		}
	}
	return -1
}

func (t *TabContainer) findTab(label string) *TabContainerTab {
	i := t.tabIndex(label)
	if i < 0 {
		return nil
	}
	return &t.Tabs[i]
}

func (t *TabContainer) removeTab(tab *TabContainerTab) {
	for i := range t.Tabs {
		if &t.Tabs[i] == tab {
			t.Tabs = slices.Delete(t.Tabs, i, i+1)
			tab.Destroy()
			if len(t.Tabs) == 0 {
				t.doc.Destroy()
				t.host.Close()
			} else if i == t.activeTab {
				if i == len(t.Tabs) {
					i--
				}
			}
			break
		}
	}
}

func (t *TabContainer) tabDragStart(e *document.Element) {
	windowing.SetDragData(t.findTab(e.Attribute("id")))
	t.host.Window.CursorSizeAll()
}

func (t *TabContainer) tabDragEnd(e *document.Element) {
	t.host.Window.CursorStandard()
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
	dd := windowing.DragData()
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
				t.selectTab(i)
				break
			}
		}
		tab.parent = weak.Make(t)
		lastParent.removeTab(tab)
		lastParent.reload()
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
		t.selectTab(to)
	}
	t.reload()
}

func (t *TabContainer) tabDragEnterRoot(e *document.Element) {
	p := e.UI.ToPanel()
	if !p.HasEnforcedColor() { // Some weird bug stacking the colors?
		e.UI.ToPanel().EnforceColor(matrix.ColorDarkGreen())
	}
}

func (t *TabContainer) tabDragLeaveRoot(e *document.Element) {
	e.UI.ToPanel().UnEnforceColor()
}

func (t *TabContainer) tabDropRoot(e *document.Element) {
	e.UI.ToPanel().UnEnforceColor()
	dd := windowing.DragData()
	tab, ok := dd.(*TabContainerTab)
	if tab == nil || !ok {
		return
	}
	lastParent := tab.parent.Value()
	tab.parent = weak.Make(t)
	if lastParent != t {
		t.Tabs = append(t.Tabs, *tab)
		t.reload()
		lastParent.removeTab(tab)
		lastParent.reload()
	} else {
		from := t.tabPtrIndex(tab)
		klib.SliceMove(t.Tabs, from, len(t.Tabs)-1)
	}
}

func (t *TabContainer) selectTab(index int) {
	if index < 0 || index >= len(t.Tabs) {
		return
	}
	if t.activeTab >= 0 {
		t.Tabs[t.activeTab].Hide()
	}
	t.Tabs[t.activeTab].Show()
}

func (t *TabContainer) tabClick(e *document.Element) {
	t.selectTab(t.tabIndex(e.Attribute("id")))
}

func (t *TabContainer) reload() {
	const html = "editor/ui/tab_container/tab_container.html"
	if t.doc != nil {
		t.doc.Destroy()
	}
	t.host.CreatingEditorEntities()
	t.doc, _ = markup.DocumentFromHTMLAsset(t.uiMan, html, t, map[string]func(*document.Element){
		"tabClick":         t.tabClick,
		"tabDragStart":     t.tabDragStart,
		"tabDragEnd":       t.tabDragEnd,
		"tabDragEnter":     t.tabDragEnter,
		"tabDragLeave":     t.tabDragLeave,
		"tabDrop":          t.tabDrop,
		"tabDragEnterRoot": t.tabDragEnterRoot,
		"tabDragLeaveRoot": t.tabDragLeaveRoot,
		"tabDropRoot":      t.tabDropRoot,
		"resizeHover":      t.resizeHover,
		"resizeExit":       t.resizeExit,
		"resizeStart":      t.resizeStart,
		"resizeStop":       t.resizeStop,
	})
	root, _ := t.doc.GetElementById("tabContent")
	t.Tabs[t.activeTab].Reload(root)
	t.host.DoneCreatingEditorEntities()
}

func newInternal(host *engine.Host, uiMan *ui.Manager, tabs []TabContainerTab) *TabContainer {
	t := &TabContainer{
		host:  host,
		uiMan: uiMan,
	}
	if t.uiMan == nil {
		t.uiMan = &ui.Manager{}
		t.uiMan.Init(host)
	}
	for i := range tabs {
		tabs[i].parent = weak.Make(t)
		t.nextTabId++
		tabs[i].Id = t.nextTabId
		t.Tabs = append(t.Tabs, tabs[i])
	}
	host.OnClose.Add(func() {
		for i := range t.Tabs {
			t.Tabs[i].Destroy()
		}
		t.doc.Destroy()
	})
	return t
}

func NewWindow(x, y int, tabs []TabContainerTab) *TabContainer {
	container := host_container.New("Kaiju Engine Tools", nil) // TODO:  Set the log stream
	go container.Run(500, 300, x, y)
	<-container.PrepLock
	t := newInternal(container.Host, nil, tabs)
	container.RunFunction(func() { t.reload() })
	return t
}

func New(host *engine.Host, uiMan *ui.Manager, tabs []TabContainerTab, snap string) *TabContainer {
	t := newInternal(host, uiMan, tabs)
	t.Snap = snap
	t.reload()
	return t
}

func (t *TabContainer) ReloadTabs(name string) {
	root, _ := t.doc.GetElementById("tabContent")
	for i := range t.Tabs {
		if t.Tabs[i].Label == name {
			t.Tabs[i].Reload(root)
		}
	}
}

func (t *TabContainer) resizeHover(e *document.Element) {
	t.host.Window.CursorSizeWE()
}

func (t *TabContainer) resizeExit(e *document.Element) {
	dd := windowing.DragData()
	if dd != t {
		t.host.Window.CursorStandard()
	}
}

func (t *TabContainer) resizeStart(e *document.Element) {
	t.host.Window.CursorSizeWE()
	windowing.SetDragData(t)
}

func (t *TabContainer) resizeStop(e *document.Element) {
	dd := windowing.DragData()
	if dd != t {
		return
	}
	t.host.Window.CursorStandard()
	//w, _ := h.doc.GetElementById("window")
	//s := w.UIPanel.Base().Layout().PixelSize().Width()
	//editor_cache.SetEditorConfigValue(sizeConfig, s)
}

func (t *TabContainer) DragUpdate() {
	win, _ := t.doc.GetElementById("window")
	x := max(50, t.host.Window.Mouse.Position().X())
	w := t.host.Window.Width()
	if int(x) < w-100 {
		win.UIPanel.Base().Layout().ScaleWidth(x)
	}
	if t.activeTab >= 0 {
		root, _ := t.doc.GetElementById("tabContent")
		t.Tabs[t.activeTab].Reload(root)
	}
}
