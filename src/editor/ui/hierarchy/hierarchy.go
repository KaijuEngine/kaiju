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
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/console"
	"kaiju/systems/events"
	"kaiju/ui"
	"strings"
)

type Hierarchy struct {
	doc        *document.Document
	input      *ui.Input
	onChangeId events.Id
}

type entityEntry struct {
	Entity *engine.Entity
}

func New() *Hierarchy {
	return &Hierarchy{}
}

func (h *Hierarchy) Destroy() {
	if h.doc != nil {
		for _, elm := range h.doc.Elements {
			elm.UI.Entity().Destroy()
		}
	}
}

func (h *Hierarchy) Create(host *engine.Host) {
	allEntities := host.Entities()
	entries := make([]entityEntry, 0, len(allEntities))
	for _, entity := range allEntities {
		entries = append(entries, entityEntry{Entity: entity})
	}
	html := klib.MustReturn(host.AssetDatabase().ReadText("ui/hierarchy/hierarchy.html"))
	h.doc = markup.DocumentFromHTMLString(host, html, "", entries, nil)
	searchInputElement, _ := h.doc.GetElementById("hierarchyInput")
	entityList, _ := h.doc.GetElementById("entityList")
	h.input = searchInputElement.UI.(*ui.Input)

	h.input.Data().OnChange.Remove(h.onChangeId)
	h.onChangeId = h.input.Data().OnChange.Add(func() {
		activeText := strings.ToLower(h.input.Text())

		for idx := range entityList.HTML.Children {
			label := entityList.HTML.Children[idx].Children[0].DocumentElement.UI.(*ui.Label)

			if strings.Contains(strings.ToLower(label.Text()), activeText) {
				entityList.HTML.Children[idx].DocumentElement.UI.Entity().Activate()
			} else {
				entityList.HTML.Children[idx].DocumentElement.UI.Entity().Deactivate()
			}
		}
	})
}

func SetupConsole(host *engine.Host) {
	hrc := New()
	console.For(host).AddCommand("hrc", func(_ *engine.Host, arg string) string {
		log := ""
		if arg == "show" {
			hrc.Destroy()
			hrc.Create(host)
		} else if arg == "hide" {
			hrc.Destroy()
		} else {
			log = "Invalid command"
		}
		return log
	})
}
