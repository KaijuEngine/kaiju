/******************************************************************************/
/* status_bar.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package status_bar

import (
	"log/slog"
	"weak"

	"kaijuengine.com/editor/common_interfaces"
	"kaijuengine.com/editor/editor_logging"
	"kaijuengine.com/editor/editor_overlay/context_menu"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/logging"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

const maxLogEntries = 100

type StatusBar struct {
	doc              *document.Document
	log              *document.Element
	logPopup         *document.Element
	logEntryTemplate *document.Element
	uiMan            ui.Manager
	logging          *editor_logging.Logging
	outerInterface   common_interfaces.Focusable
	inPopup          bool
}

func (b *StatusBar) Initialize(host *engine.Host, logging *editor_logging.Logging, outerInterface common_interfaces.Focusable) error {
	defer tracing.NewRegion("StatusBar.Initialize").End()
	b.logging = logging
	b.outerInterface = outerInterface
	b.uiMan.Init(host)
	var err error
	b.doc, err = markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/status_bar.go.html",
		nil, map[string]func(*document.Element){
			"openLogWindow":      b.openLogWindow,
			"closePopup":         b.closePopup,
			"rightClickLogEntry": b.rightClickLogEntry,
		})
	b.setupUIReferences()
	b.bindToSlog()
	return err
}

func (b *StatusBar) Focus() { b.uiMan.EnableUpdate() }
func (b *StatusBar) Blur() {
	defer tracing.NewRegion("StatusBar.Blur").End()
	if b.inPopup {
		return
	}
	b.uiMan.DisableUpdate()
}

func (b *StatusBar) IsFocusedOnInput() bool {
	return b.uiMan.Group.IsFocusedOnInput()
}

func (b *StatusBar) setupUIReferences() {
	defer tracing.NewRegion("StatusBar.setupUIReferences").End()
	b.log, _ = b.doc.GetElementById("log")
	b.logPopup, _ = b.doc.GetElementById("logPopup")
	b.logEntryTemplate, _ = b.doc.GetElementById("logEntryTemplate")
	b.logPopup.UI.Hide()
}

func (b *StatusBar) bindToSlog() {
	defer tracing.NewRegion("StatusBar.bindToSlog").End()
	wb := weak.Make(b)
	b.logging.OnNewLog = func(msg editor_logging.Message) {
		bar := wb.Value()
		if bar == nil {
			return
		}
		b.uiMan.Host.RunOnMainThread(func() {
			bar.doc.SetElementClasses(bar.log, msg.Category+"Status")
			bar.setLog(msg.Message)
			elm := b.doc.DuplicateElement(b.logEntryTemplate)
			elm.Children[0].UI.ToLabel().SetText(msg.ToString())
			b.doc.SetElementClassesWithoutApply(elm, "logLine", msg.Category)
			parent := elm.Parent.Value()
			if len(parent.Children) > maxLogEntries {
				// +1 because template is 0
				for i := 1; i <= len(parent.Children)-maxLogEntries; i++ {
					b.doc.RemoveElement(parent.Children[i])
				}
			}
		})
	}
}

func (b *StatusBar) setLog(msg string) {
	defer tracing.NewRegion("StatusBar.setLog").End()
	res := logging.ToMap(msg)
	lbl := b.log.Children[0].UI.ToLabel()
	if m, ok := res["msg"]; ok {
		lbl.SetText(m)
	} else {
		lbl.SetText(msg)
	}
}

func (b *StatusBar) openLogWindow(*document.Element) {
	defer tracing.NewRegion("StatusBar.openLogWindow").End()
	if b.inPopup {
		return
	}
	b.inPopup = true
	b.logPopup.UI.Show()
	b.logEntryTemplate.UI.Hide()
	pnl := b.logPopup.UI.ToPanel()
	b.uiMan.Host.RunAfterFrames(2, func() {
		pnl.SetScrollY(pnl.MaxScroll().Y())
	})
	b.outerInterface.BlurInterface()
}

func (b *StatusBar) closePopup(*document.Element) {
	defer tracing.NewRegion("StatusBar.closePopup").End()
	b.logPopup.UI.Hide()
	b.outerInterface.FocusInterface()
	b.inPopup = false
}

func (b *StatusBar) rightClickLogEntry(e *document.Element) {
	defer tracing.NewRegion("StatusBar.rightClickLogEntry").End()
	if len(e.Children) == 0 {
		return
	}
	text := e.Children[0].UI.ToLabel().Text()
	options := []context_menu.ContextMenuOption{
		{
			Label: "Copy to clipboard",
			Call:  func() { b.uiMan.Host.Window.CopyToClipboard(text) },
		},
		{
			Label: "Open log file",
			Call: func() {
				path, err := logging.LogFilePath()
				if err != nil {
					slog.Error("failed to locate the log file path", "error", err)
					return
				}
				if err := filesystem.OpenFileInTextEditor(path); err != nil {
					slog.Error("failed to open the log file",
						"path", path, "error", err)
				}
			},
		},
	}
	pos := b.uiMan.Host.Window.Mouse.ScreenPosition()
	context_menu.Show(b.uiMan.Host, options, pos, func() {})
}
