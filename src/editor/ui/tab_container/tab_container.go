/******************************************************************************/
/* tab_container.go                                                           */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package tab_container

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/engine/host_container"
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/windowing"
	"kaiju/rendering"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"weak"

	"github.com/KaijuEngine/uuid"
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
	host           weak.Pointer[engine.Host]
	doc            *document.Document
	activeTab      int
	uiMan          *ui.Manager
	Tabs           []TabContainerTab
	Snap           string
	dragScale      float32
	usingOwnWindow bool
}

func (t *TabContainer) tabPtrIndex(tab *TabContainerTab) int {
	for i := range t.Tabs {
		if &t.Tabs[i] == tab {
			return i
		}
	}
	return -1
}

func (t *TabContainer) tabIndex(id string) int {
	for i := range t.Tabs {
		if t.Tabs[i].Id == id {
			return i
		}
	}
	return -1
}

func (t *TabContainer) findTab(id string) *TabContainerTab {
	i := t.tabIndex(id)
	if i < 0 {
		return nil
	}
	return &t.Tabs[i]
}

func (t *TabContainer) removeTab(tab *TabContainerTab) {
	host := t.host.Value()
	debug.EnsureNotNil(host)
	for i := range t.Tabs {
		if &t.Tabs[i] == tab {
			tab.Destroy()
			t.Tabs = slices.Delete(t.Tabs, i, i+1)
			if len(t.Tabs) == 0 {
				t.doc.Destroy()
				t.doc = nil
				t.activeTab = -1
				if t.usingOwnWindow {
					host.Close()
				}
			} else if t.activeTab >= len(t.Tabs) {
				t.activeTab = len(t.Tabs) - 1
			}
			break
		}
	}
}

func (t *TabContainer) tabDragStart(e *document.Element) {
	windowing.SetDragData(t.findTab(e.Attribute("id")))
	host := t.host.Value()
	debug.EnsureNotNil(host)
	host.Window.CursorSizeAll()
	// We don't need to remove this because the event clears after removal
	windowing.OnDragStop.Add(func() { t.tabDragEnd(e) })
}

func (t *TabContainer) tabDragEnd(e *document.Element) {
	host := t.host.Value()
	debug.EnsureNotNil(host)
	if !windowing.IsDragDataUsed() {
		dd := windowing.UseDragData()
		tab, ok := dd.(*TabContainerTab)
		if !ok {
			return
		}
		sp := host.Window.Mouse.ScreenPosition()
		tab.Destroy()
		copy := *tab
		copy.parent = weak.Pointer[TabContainer]{}
		t.removeTab(tab)
		t.reload()
		x, y := host.Window.ToScreenPosition(int(sp.X()), int(sp.Y()))
		NewWindow(copy.Label, x, y, []TabContainerTab{copy}, host.LogStream)
	}
	host.Window.CursorStandard()
}

func (t *TabContainer) tabDragEnter(e *document.Element) {
	host := t.host.Value()
	debug.EnsureNotNil(host)
	tex, err := host.TextureCache().Texture("textures/window_tab_drag_enter.png",
		rendering.TextureFilterNearest)
	if err == nil {
		e.UI.ToPanel().SetBackground(tex)
	}
}

func (t *TabContainer) tabDragLeave(e *document.Element) {
	host := t.host.Value()
	debug.EnsureNotNil(host)
	tex, err := host.TextureCache().Texture("textures/window_tab.png",
		rendering.TextureFilterNearest)
	if err == nil {
		e.UI.ToPanel().SetBackground(tex)
	}
}

func (t *TabContainer) calcTabWidths() {
	tabElms := t.doc.GetElementsByGroup("tabGroup")
	for i := range tabElms {
		lbl := tabElms[i].Children[0].UI.ToLabel()
		width := int(lbl.Measure().X() + 8)
		tabElms[i].UI.ToPanel().Base().Layout().ScaleWidth(float32(width))
		// The tab has a CSS hover effect so we need to alter the styling
		for j := range tabElms[i].StyleRules {
			if tabElms[i].StyleRules[j].Property == "width" {
				tabElms[i].StyleRules[j].Values[0].Str = strconv.Itoa(width) + "px"
			}
		}
	}
}

func (t *TabContainer) resetTabTextures() {
	// Fixes drop bug running to fast and swapping the texture to drag hover
	host := t.host.Value()
	debug.EnsureNotNil(host)
	host.RunAfterFrames(2, func() {
		tabElms := t.doc.GetElementsByGroup("tabGroup")
		for i := range tabElms {
			t.tabDragLeave(tabElms[i])
		}
		inner, _ := t.doc.GetElementById("tabsInner")
		inner.UI.ToPanel().UnEnforceColor()
	})
}

func (t *TabContainer) tabDrop(e *document.Element) {
	dd := windowing.UseDragData()
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
		t.reload()
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
	t.resetTabTextures()
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
	dd := windowing.UseDragData()
	tab, ok := dd.(*TabContainerTab)
	if tab == nil || !ok {
		return
	}
	lastParent := tab.parent.Value()
	tab.parent = weak.Make(t)
	if lastParent != t {
		t.Tabs = append(t.Tabs, *tab)
		if lastParent != nil {
			lastParent.removeTab(tab)
			if len(lastParent.Tabs) > 0 {
				lastParent.reload()
			}
		}
	} else {
		if from := t.tabPtrIndex(tab); from >= 0 {
			klib.SliceMove(t.Tabs, from, len(t.Tabs)-1)
		} else {
			slog.Error("there was an issue with tab repartinging, now recovering...")
		}
	}
	t.selectTab(len(t.Tabs) - 1)
	t.resetTabTextures()
}

func (t *TabContainer) selectTab(index int) {
	if index < 0 || index >= len(t.Tabs) || index == t.activeTab {
		return
	}
	t.activeTab = index
	t.reload()
}

func (t *TabContainer) tabClick(e *document.Element) {
	t.selectTab(t.tabIndex(e.Attribute("id")))
}

func (t *TabContainer) reload() {
	const html = "editor/ui/tab_container/tab_container.html"
	for i := range t.Tabs {
		t.Tabs[i].Destroy()
	}
	if t.doc != nil {
		t.doc.Destroy()
	}
	host := t.host.Value()
	debug.EnsureNotNil(host)
	host.CreatingEditorEntities()
	t.doc, _ = markup.DocumentFromHTMLAsset(t.uiMan, html, t, map[string]func(*document.Element){
		"tabClick":         t.tabClick,
		"tabDragStart":     t.tabDragStart,
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
	if len(t.Tabs) > 0 {
		klib.Clamp(t.activeTab, 0, len(t.Tabs)-1)
		root, _ := t.doc.GetElementById("tabContent")
		t.Tabs[t.activeTab].Reload(t.uiMan, root)
	}
	host.DoneCreatingEditorEntities()
	if t.dragScale > 0 {
		win, _ := t.doc.GetElementById("window")
		switch t.Snap {
		case "left":
			fallthrough
		case "right":
			win.UIPanel.Base().Layout().ScaleWidth(t.dragScale)
		case "top":
			fallthrough
		case "bottom":
			win.UIPanel.Base().Layout().ScaleHeight(t.dragScale)
		}
	}
	t.calcTabWidths()
	t.doc.Clean()
}

func newInternal(host *engine.Host, uiMan *ui.Manager, tabs []TabContainerTab) *TabContainer {
	t := &TabContainer{
		host:  weak.Make(host),
		uiMan: uiMan,
	}
	if t.uiMan == nil {
		t.uiMan = &ui.Manager{}
		t.uiMan.Init(host)
	}
	for i := range tabs {
		tabs[i].parent = weak.Make(t)
		if tabs[i].Id == "" {
			tabs[i].Id = uuid.New().String()
		}
		t.Tabs = append(t.Tabs, tabs[i])
	}
	host.OnClose.Add(func() {
		for i := range t.Tabs {
			t.Tabs[i].Destroy()
		}
		if t.doc != nil {
			t.doc.Destroy()
		}
	})
	return t
}

func NewWindow(title string, x, y int, tabs []TabContainerTab, logStream *logging.LogStream) *TabContainer {
	const w = 500
	const h = 300
	container := host_container.New(title, logStream)
	x = klib.Clamp(x-w/2, 0, math.MaxInt32)
	y = klib.Clamp(y-h/2, 0, math.MaxInt32)
	go container.Run(w, h, x, y)
	<-container.PrepLock
	t := newInternal(container.Host, nil, tabs)
	t.usingOwnWindow = true
	container.RunFunction(func() { t.reload() })
	container.Host.Window.Focus()
	t.selectTab(0)
	return t
}

func New(host *engine.Host, uiMan *ui.Manager, tabs []TabContainerTab, snap string) *TabContainer {
	t := newInternal(host, uiMan, tabs)
	t.Snap = snap
	t.reload()
	return t
}

func (t *TabContainer) ReloadTabs(name string, forceOpen bool) bool {
	found := false
	for i := range t.Tabs {
		if t.Tabs[i].Label == name {
			if (t.doc != nil && t.doc.IsActive()) || forceOpen {
				if t.activeTab == i {
					t.reload()
				} else {
					t.selectTab(i)
				}
				found = true
			}
		}
	}
	return found
}

func (t *TabContainer) resizeHover(e *document.Element) {
	host := t.host.Value()
	debug.EnsureNotNil(host)
	switch t.Snap {
	case "left":
		fallthrough
	case "right":
		host.Window.CursorSizeWE()
	case "top":
		fallthrough
	case "bottom":
		host.Window.CursorSizeNS()
	}
}

func (t *TabContainer) resizeExit(e *document.Element) {
	dd := windowing.DragData()
	if dd != t {
		host := t.host.Value()
		debug.EnsureNotNil(host)
		host.Window.CursorStandard()
	}
}

func (t *TabContainer) resizeStart(e *document.Element) {
	host := t.host.Value()
	debug.EnsureNotNil(host)
	switch t.Snap {
	case "left":
		fallthrough
	case "right":
		host.Window.CursorSizeWE()
	case "top":
		fallthrough
	case "bottom":
		host.Window.CursorSizeNS()
	}
	windowing.SetDragData(t)
}

func (t *TabContainer) resizeStop(e *document.Element) {
	dd := windowing.DragData()
	if dd != t {
		return
	}
	host := t.host.Value()
	debug.EnsureNotNil(host)
	host.Window.CursorStandard()
	//w, _ := h.doc.GetElementById("window")
	//s := w.UIPanel.Base().Layout().PixelSize().Width()
	//editor_cache.SetEditorConfigValue(sizeConfig, s)
}

func (t *TabContainer) Hide() {
	t.doc.Deactivate()
}

func (t *TabContainer) Show() {
	if t.doc == nil {
		t.reload()
	} else {
		t.doc.Activate()
	}
}

func (t *TabContainer) Toggle() {
	if t.doc == nil || !t.doc.Elements[0].UI.Entity().IsActive() {
		t.Show()
	} else {
		t.Hide()
	}
}

func (t *TabContainer) DragUpdate() {
	win, _ := t.doc.GetElementById("window")
	host := t.host.Value()
	debug.EnsureNotNil(host)
	switch t.Snap {
	case "left":
		w := float32(host.Window.Width())
		t.dragScale = klib.Clamp(host.Window.Mouse.Position().X(), 50, w-100)
		win.UIPanel.Base().Layout().ScaleWidth(t.dragScale)
	case "top":
		h := float32(host.Window.Height())
		t.dragScale = klib.Clamp(matrix.Float(h)-host.Window.Mouse.Position().Y(), 50, h-100)
		win.UIPanel.Base().Layout().ScaleHeight(t.dragScale)
	case "bottom":
		h := float32(host.Window.Height())
		t.dragScale = klib.Clamp(host.Window.Mouse.Position().Y()-20, 50, h-100)
		win.UIPanel.Base().Layout().ScaleHeight(t.dragScale)
	case "right":
		w := float32(host.Window.Width())
		t.dragScale = klib.Clamp(matrix.Float(w)-host.Window.Mouse.Position().X(), 50, w-100)
		win.UIPanel.Base().Layout().ScaleWidth(t.dragScale)
	case "center":
		// This shouldn't be something that happens...
	}
	if t.activeTab >= 0 {
		root, _ := t.doc.GetElementById("tabContent")
		t.Tabs[t.activeTab].Reload(t.uiMan, root)
	}
}
