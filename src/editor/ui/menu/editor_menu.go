/******************************************************************************/
/* editor_menu.go                                                             */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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

package menu

import (
	"kaiju/editor/content/content_opener"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/about_window"
	"kaiju/editor/ui/content_window"
	"kaiju/editor/ui/log_window"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"log/slog"
	"os/exec"
	"runtime"
)

type Menu struct {
	container     *host_container.Container
	doc           *document.Document
	isOpen        bool
	logWindow     *log_window.LogWindow
	contentOpener *content_opener.Opener
	editor        interfaces.Editor
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

func openAbout(*document.DocElement) {
	// TODO:  Open the about in a new window
	about_window.New()
}

func openRepository(*document.DocElement) {
	cmd := "open"
	if runtime.GOOS == "windows" {
		cmd = "explorer"
	}
	exec.Command(cmd, "https://github.com/KaijuEngine/kaiju").Run()
}

func (m *Menu) openLogWindow(*document.DocElement) {
	m.logWindow.Show()
}

func (m *Menu) openContentWindow(*document.DocElement) {
	content_window.New(m.contentOpener, m.editor)
}

func (m *Menu) saveStage(*document.DocElement) {
	if err := m.editor.StageManager().Save(); err != nil {
		slog.Error("Save stage failed", slog.String("error", err.Error()))
	}
}

func New(container *host_container.Container,
	logWindow *log_window.LogWindow,
	contentOpener *content_opener.Opener,
	editor interfaces.Editor) *Menu {

	host := container.Host
	html := klib.MustReturn(host.AssetDatabase().ReadText("editor/ui/menu.html"))
	m := &Menu{
		container:     container,
		logWindow:     logWindow,
		contentOpener: contentOpener,
		editor:        editor,
	}
	funcMap := map[string]func(*document.DocElement){
		"openLogWindow":     m.openLogWindow,
		"openRepository":    openRepository,
		"openAbout":         openAbout,
		"saveStage":         m.saveStage,
		"openContentWindow": m.openContentWindow,
		"sampleInfo":        func(*document.DocElement) { slog.Info("This is some info") },
		"sampleWarn":        func(*document.DocElement) { slog.Warn("This is a warning") },
		"sampleError":       func(*document.DocElement) { slog.Error("This is an error") },
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
