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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor_menu

import (
	"kaiju/editor/content/content_opener"
	"kaiju/editor/editor_interface"
	"kaiju/editor/ui/about_window"
	"kaiju/editor/ui/content_window"
	"kaiju/editor/ui/editor_settings_window"
	"kaiju/editor/ui/hierarchy"
	"kaiju/editor/ui/log_window"
	"kaiju/editor/ui/shader_designer"
	"kaiju/engine"
	"kaiju/engine/host_container"
	"kaiju/engine/systems/console"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/audio"
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
	editor          editor_interface.Editor
	uiGroup         *ui.Group
	uiMan           *ui.Manager
}

func New(container *host_container.Container,
	logWindow *log_window.LogWindow,
	contentWindow *content_window.ContentWindow,
	hierarchyWindow *hierarchy.Hierarchy,
	contentOpener *content_opener.Opener,
	editor editor_interface.Editor,
	uiMan *ui.Manager) *Menu {

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
		"openLogWindow":            m.openLogWindow,
		"openRepository":           openRepository,
		"openAbout":                m.openAbout,
		"newStage":                 m.newStage,
		"saveStage":                m.saveStage,
		"openProject":              m.openProject,
		"openContentWindow":        m.openContentWindow,
		"openHierarchyWindow":      m.openHierarchyWindow,
		"openShaderDesignerWindow": m.openShaderDesignerWindow,
		"newEntity":                m.newEntity,
		"newCone":                  m.newCone,
		"newCube":                  m.newCube,
		"newCylinder":              m.newCylinder,
		"newIcoSphere":             m.newIcoSphere,
		"newPlane":                 m.newPlane,
		"newSphere":                m.newSphere,
		"newTorus":                 m.newTorus,
		"showEditorSettings":       m.showEditorSettings,
	}
	m.doc = markup.DocumentFromHTMLString(uiMan, html, "", nil, funcMap, nil)
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
	about_window.New()
}

func openRepository(*document.Element) {
	klib.OpenWebsite("https://github.com/KaijuEngine/kaiju")
}

func (m *Menu) openLogWindow(*document.Element) {
	m.editor.ReloadOrOpenTab("Log")
}

func (m *Menu) openContentWindow(*document.Element) {
	m.editor.ReloadOrOpenTab("Content")
}

func (m *Menu) openHierarchyWindow(*document.Element) {
	m.editor.ReloadOrOpenTab("Hierarchy")
}

func (m *Menu) openShaderDesignerWindow(*document.Element) {
	shader_designer.New(shader_designer.StateHome, m.container.Host.LogStream)
}

func (m *Menu) newEntity(*document.Element) {
	m.editor.CreateEntity("Entity")
}

func (m *Menu) newCone(*document.Element)      { m.createCone() }
func (m *Menu) newCube(*document.Element)      { m.createCube() }
func (m *Menu) newCylinder(*document.Element)  { m.createCylinder() }
func (m *Menu) newIcoSphere(*document.Element) { m.createIcoSphere() }
func (m *Menu) newPlane(*document.Element)     { m.createPlane() }
func (m *Menu) newSphere(*document.Element)    { m.createSphere() }
func (m *Menu) newTorus(*document.Element)     { m.createTorus() }

func (m *Menu) showEditorSettings(*document.Element) {
	editor_settings_window.New()
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
		wav := audio.NewClip("editor/audio/sfx/fanfare.wav")
		host.Audio().Play(wav)
		return "Playing fanfare.wav"
	})
}
