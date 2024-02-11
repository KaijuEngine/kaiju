/*****************************************************************************/
/* editor_menu.go                                                            */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

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
