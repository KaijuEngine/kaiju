/******************************************************************************/
/* editor_menu.go                                                             */
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

package menu

import (
	"kaiju/audio/audio_system"
	"kaiju/editor/content/content_opener"
	"kaiju/editor/interfaces"
	"kaiju/editor/ui/about_window"
	"kaiju/editor/ui/content_window"
	"kaiju/editor/ui/hierarchy"
	"kaiju/editor/ui/log_window"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/console"
	"kaiju/ui"
	"log/slog"
)

type Menu struct {
	container       *host_container.Container
	doc             *document.Document
	isOpen          bool
	logWindow       *log_window.LogWindow
	contentWindow   *content_window.ContentWindow
	hierarchyWindow *hierarchy.Hierarchy
	contentOpener   *content_opener.Opener
	editor          interfaces.Editor
	uiGroup         *ui.Group
}

func New(container *host_container.Container,
	logWindow *log_window.LogWindow,
	contentWindow *content_window.ContentWindow,
	hierarchyWindow *hierarchy.Hierarchy,
	contentOpener *content_opener.Opener,
	editor interfaces.Editor,
	uiGroup *ui.Group) *Menu {

	host := container.Host
	html := klib.MustReturn(host.AssetDatabase().ReadText("editor/ui/menu.html"))
	m := &Menu{
		container:       container,
		logWindow:       logWindow,
		contentWindow:   contentWindow,
		hierarchyWindow: hierarchyWindow,
		contentOpener:   contentOpener,
		editor:          editor,
	}
	funcMap := map[string]func(*document.Element){
		"openLogWindow":       m.openLogWindow,
		"openRepository":      openRepository,
		"openAbout":           m.openAbout,
		"newStage":            m.newStage,
		"saveStage":           m.saveStage,
		"openProject":         m.openProject,
		"openContentWindow":   m.openContentWindow,
		"openHierarchyWindow": m.openHierarchyWindow,
		"newEntity":           m.newEntity,
	}
	m.doc = markup.DocumentFromHTMLString(host, html, "", nil, funcMap)
	m.doc.SetGroup(uiGroup)
	allItems := m.doc.GetElementsByClass("menuItem")
	for i := range allItems {
		targetId := allItems[i].Attribute("data-target")
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
				// TODO:  The CSS isn't working here?
				l.SetZ(5)
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
	m.setupConsoleCommands()
	return m
}

func (m *Menu) close() {
	all := m.doc.GetElementsByClass("menuItemList")
	for i := range all {
		all[i].UI.Entity().Deactivate()
	}
}

func (m *Menu) open(target *document.Element) {
	m.close()
	target.UI.Entity().SetActive(m.isOpen)
}

func (m *Menu) openMenu(targetId string) {
	if t, ok := m.doc.GetElementById(targetId); ok {
		m.isOpen = !m.isOpen
		m.open(t)
	}
}

func (m *Menu) hoverOpenMenu(targetId string) {
	if !m.isOpen {
		return
	}
	if t, ok := m.doc.GetElementById(targetId); ok {
		m.open(t)
	}
}

func (m *Menu) openAbout(*document.Element) {
	// TODO:  Open the about in a new window
	about_window.New(m.editor.Host().AssetDatabase().EditorContext.EditorPath)
}

func openRepository(*document.Element) {
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju")
}

func (m *Menu) openLogWindow(*document.Element) {
	m.logWindow.Show()
}

func (m *Menu) openContentWindow(*document.Element) {
	m.contentWindow.Show()
}

func (m *Menu) openHierarchyWindow(*document.Element) {
	m.hierarchyWindow.Show()
}

func (m *Menu) newEntity(*document.Element) {
	m.editor.CreateEntity("Entity")
}

func (m *Menu) newStage(*document.Element) {
	m.editor.StageManager().New()
}

func (m *Menu) saveStage(*document.Element) {
	m.editor.StageManager().Save(m.editor.StatusBar())
}

func (m *Menu) openProject(*document.Element) {
	m.editor.OpenProject()
}

func (m *Menu) setupConsoleCommands() {
	c := console.For(m.editor.Host())
	c.AddCommand("log", "Opens the log window", func(_ *engine.Host, arg string) string {
		switch arg {
		case "info":
			slog.Info("This is some info")
			return "Generated a sample info message"
		case "warn":
			slog.Warn("This is a warning")
			return "Generated a sample warning message"
		case "error":
			slog.Error("This is an error")
			return "Generated a sample error message"
		default:
			m.openLogWindow(nil)
			return ""
		}
	})
	c.AddCommand("content", "Opens a content window", func(*engine.Host, string) string {
		m.openContentWindow(nil)
		return ""
	})

	c.AddCommand("audio.test", "Tests playback of a wav", func(host *engine.Host, _ string) string {
		wav, err := audio_system.LoadWav(host.AssetDatabase(), "editor/audio/sfx/fanfare.wav")
		if err != nil {
			return err.Error()
		}
		host.Audio().Play(wav)
		return "Playing fanfare.wav"
	})
}
