/******************************************************************************/
/* hierarchy.go                                                               */
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

package hierarchy

import (
	"kaiju/editor/selection"
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/ui/drag_datas"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/events"
	"kaiju/ui"
	"kaiju/windowing"
	"log/slog"
	"strings"
)

const sizeConfig = "hierarchyWindowSize"

type Hierarchy struct {
	host                 *engine.Host
	entityCtxMenuActions []context_menu.ContextMenuEntry
	ctxMenuSet           context_menu.ContextMenuSet
	selection            *selection.Selection
	doc                  *document.Document
	input                *ui.Input
	searchText           string
	query                string
	reloadTab            func(name string)
	focusOnReload        bool
}

func (h *Hierarchy) Document() *document.Document { return h.doc }
func (h *Hierarchy) TabTitle() string             { return "Hierarchy" }

func (h *Hierarchy) Destroy() {
	if h.doc != nil {
		h.focusOnReload = h.input.IsFocused()
		h.doc.Destroy()
		h.doc = nil
	}
}

type entityEntry struct {
	Entity          *engine.Entity
	ShowingChildren bool
}

type hierarchyData struct {
	Entries    []entityEntry
	SearchText string
	Query      string
}

func (e entityEntry) Depth() int {
	depth := 0
	p := e.Entity
	for p.Parent != nil {
		depth++
		p = p.Parent
	}
	return depth
}

func New(host *engine.Host, selection *selection.Selection,
	ctxMenuSet context_menu.ContextMenuSet, reloadTabFunc func(name string)) *Hierarchy {
	h := &Hierarchy{
		host:       host,
		selection:  selection,
		ctxMenuSet: ctxMenuSet,
		reloadTab:  reloadTabFunc,
	}
	h.selection.Changed.Add(h.onSelectionChanged)
	return h
}

func (h *Hierarchy) orderEntitiesVisually() []entityEntry {
	allEntities := h.host.Entities()
	entries := make([]entityEntry, 0, len(allEntities))
	roots := make([]*engine.Entity, 0, len(allEntities))
	for _, entity := range allEntities {
		if entity.IsRoot() && !entity.EditorBindings.IsDeleted {
			roots = append(roots, entity)
		}
	}
	var addChildren func(*engine.Entity)
	addChildren = func(entity *engine.Entity) {
		if entity.EditorBindings.IsDeleted {
			return
		}
		entries = append(entries, entityEntry{entity, false})
		for _, c := range entity.Children {
			addChildren(c)
		}
	}
	for _, r := range roots {
		addChildren(r)
	}
	return entries
}

func (h *Hierarchy) filter(entries []entityEntry) []entityEntry {
	if h.query == "" {
		return entries
	}
	filtered := make([]entityEntry, 0, len(entries))
	// TODO:  Append the entire path to the entity kf not already appended
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Entity.Name()), h.query) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func (h *Hierarchy) Reload(uiMan *ui.Manager, root *document.Element) {
	focusInput := h.focusOnReload
	if h.doc != nil {
		h.focusOnReload = h.input.IsFocused()
		h.doc.Destroy()
	}
	data := hierarchyData{
		Entries:    h.filter(h.orderEntitiesVisually()),
		SearchText: h.searchText,
		Query:      h.query,
	}
	host := h.host
	host.CreatingEditorEntities()
	h.doc = klib.MustReturn(markup.DocumentFromHTMLAssetRooted(
		uiMan, "editor/ui/hierarchy_window.html", data,
		map[string]func(*document.Element){
			"selectedEntity": h.selectedEntity,
			"dragStart":      h.dragStart,
			"drop":           h.drop,
			"dragEnter":      h.dragEnter,
			"dragExit":       h.dragExit,
			"entryCtxMenu":   h.entryCtxMenu,
			"search":         h.submitSearch,
		}, root))
	host.DoneCreatingEditorEntities()
	if elm, ok := h.doc.GetElementById("searchInput"); !ok {
		slog.Error(`Failed to locate the "searchInput" for the hierarchy`)
	} else {
		h.input = elm.UI.ToInput()
	}
	h.doc.Clean()
	if focusInput {
		h.input.Focus()
	}
}

func (h *Hierarchy) submit() {
	h.searchText = h.input.Text()
	h.query = strings.ToLower(strings.TrimSpace(h.input.Text()))
	h.reloadTab(h.TabTitle())
}

func (h *Hierarchy) onSelectionChanged() {
	if h.doc == nil {
		return
	}
	elm, ok := h.doc.GetElementById("list")
	if !ok {
		slog.Error("Could not find hierarchy list, reopen the hierarchy window")
		return
	}
	for i := range elm.Children {
		elm.Children[i].UnEnforceColor()
	}
	for _, c := range elm.Children {
		id := engine.EntityId(c.Attribute("id"))
		for _, se := range h.selection.Entities() {
			if se.Id() == id {
				c.EnforceColor(matrix.ColorDarkBlue())
				break
			}
		}
	}
}

func (h *Hierarchy) selectedEntity(elm *document.Element) {
	id := engine.EntityId(elm.Attribute("id"))
	if e, ok := h.host.FindEntity(id); !ok {
		slog.Error("Could not find entity", slog.String("id", string(id)))
	} else {
		kb := &h.host.Window.Keyboard
		if kb.HasCtrl() {
			h.selection.Toggle(e)
		} else if kb.HasShift() {
			h.selection.Add(e)
		} else {
			h.selection.Set(e)
		}
		h.onSelectionChanged()
	}
}

func (h *Hierarchy) drop(elm *document.Element) {
	elm.UnEnforceColor()
	from, ok := windowing.DragData().(*drag_datas.EntityIdDragData)
	if !ok {
		return
	}
	windowing.UseDragData()
	if f, ok := h.host.FindEntity(from.EntityId); ok {
		toId := elm.Attribute("id")
		if toId != "" {
			to := engine.EntityId(toId)
			if t, ok := h.host.FindEntity(to); ok {
				f.SetParent(t)
				h.reloadTab(h.TabTitle())
			} else {
				slog.Error("Could not find drop target entity", slog.String("id", string(to)))
			}
		} else {
			f.SetParent(nil)
			h.reloadTab(h.TabTitle())
		}
	} else {
		slog.Error("Could not find drag entity", slog.String("id", string(from.EntityId)))
	}
}

func (h *Hierarchy) dragStart(elm *document.Element) {
	id := engine.EntityId(elm.Attribute("id"))
	h.host.Window.CursorSizeAll()
	windowing.SetDragData(&drag_datas.EntityIdDragData{id})
	elm.EnforceColor(matrix.ColorPurple())
	var eid events.Id
	eid = windowing.OnDragStop.Add(func() {
		h.host.Window.CursorStandard()
		windowing.OnDragStop.Remove(eid)
		elm.UnEnforceColor()
	})
}

func (h *Hierarchy) dragEnter(elm *document.Element) {
	myId := engine.EntityId(elm.Attribute("id"))
	if dd, ok := windowing.DragData().(*drag_datas.EntityIdDragData); !ok {
		return
	} else {
		if myId != dd.EntityId {
			elm.EnforceColor(matrix.ColorOrange())
		}
	}
}

func (h *Hierarchy) dragExit(elm *document.Element) {
	myId := engine.EntityId(elm.Attribute("id"))
	if dd, ok := windowing.DragData().(*drag_datas.EntityIdDragData); !ok {
		return
	} else {
		if myId != dd.EntityId {
			elm.UnEnforceColor()
		}
	}
}

func (h *Hierarchy) entryCtxMenu(elm *document.Element) {
	eid := engine.EntityId(elm.Attribute("id"))
	e, ok := h.host.FindEntity(eid)
	if !ok {
		return
	}
	if !h.selection.Contains(e) {
		h.selection.Set(e)
	}
	h.ctxMenuSet.Show()
}

func (h *Hierarchy) submitSearch(*document.Element) {
	h.submit()
}
