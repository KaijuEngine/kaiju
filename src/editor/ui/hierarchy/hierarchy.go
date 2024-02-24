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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package hierarchy

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/editor_window"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/events"
	"kaiju/ui"
	"log/slog"
)

type Hierarchy struct {
	editor     interfaces.Editor
	container  *host_container.Container
	doc        *document.Document
	input      *ui.Input
	onChangeId events.Id
}

type entityEntry struct {
	Entity *engine.Entity
}

func (h *Hierarchy) Tag() string                          { return editor_cache.HierarchyWindow }
func (h *Hierarchy) Container() *host_container.Container { return h.container }
func (h *Hierarchy) Closed()                              {}

func New(editor interfaces.Editor) {
	h := &Hierarchy{
		editor:    editor,
		container: host_container.New("Hierarchy", nil),
	}
	editor_window.OpenWindow(h, 300, 600, -1, -1)
	allEntities := editor.Host().Entities()
	entries := make([]entityEntry, 0, len(allEntities))
	for _, entity := range allEntities {
		entries = append(entries, entityEntry{Entity: entity})
	}
	h.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(
		h.container.Host, "editor/ui/hierarchy_window.html", entries,
		map[string]func(*document.DocElement){
			"selectedEntity": h.onSelectedEntity,
		}))
}

func (h *Hierarchy) onSelectedEntity(elm *document.DocElement) {
	id := elm.HTML.Attribute("id")
	if e, ok := h.editor.Host().FindEntity(id); !ok {
		slog.Error("Could not find entity", slog.String("id", id))
	} else {
		kb := &h.container.Host.Window.Keyboard
		if kb.HasCtrl() {
			h.editor.Selection().Toggle(e)
		} else if kb.HasShift() {
			h.editor.Selection().Add(e)
		} else {
			h.editor.Selection().Set(e)
		}
		for i := range elm.HTML.Parent.Children {
			elm.HTML.Parent.Children[i].DocumentElement.UIPanel.UnEnforceColor()
		}
		for i := range elm.HTML.Parent.Children {
			child := &elm.HTML.Parent.Children[i]
			id := child.Parent.Children[i].Attribute("id")
			for _, se := range h.editor.Selection().Entities() {
				if se.Id() == id {
					child.DocumentElement.UIPanel.EnforceColor(matrix.ColorDarkBlue())
					break
				}
			}
		}
	}
}
