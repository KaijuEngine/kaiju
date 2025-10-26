/******************************************************************************/
/* status_bar.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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

package status_bar

import (
	"kaiju/editor/common_interfaces"
	"kaiju/editor/editor_logging"
	"kaiju/engine"
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"regexp"
	"strconv"
	"strings"
	"weak"
)

type StatusBar struct {
	doc              *document.Document
	msg              *document.Element
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
			"openLogWindow": b.openLogWindow,
			"closePopup":    b.closePopup,
		})
	b.setupUIReferences()
	b.bindToSlog()
	return err
}

func (b *StatusBar) Focus() { b.uiMan.EnableUpdate() }
func (b *StatusBar) Blur() {
	if b.inPopup {
		return
	}
	b.uiMan.DisableUpdate()
}

func (b *StatusBar) setupUIReferences() {
	defer tracing.NewRegion("StatusBar.setupUIReferences").End()
	b.msg, _ = b.doc.GetElementById("msg")
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
		bar.doc.SetElementClasses(bar.log, msg.Category+"Status")
		bar.setLog(msg.Message)
		elm := b.doc.DuplicateElement(b.logEntryTemplate)
		elm.Children[0].UI.ToLabel().SetText(msg.ToString())
		b.doc.SetElementClassesWithoutApply(elm, "logLine", msg.Category)
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

func (b *StatusBar) SetMessage(status string) {
	defer tracing.NewRegion("StatusBar.SetMessage").End()
	lbl := b.msg.Children[0].UI.ToLabel()
	t := lbl.Text()
	if strings.HasSuffix(t, status) {
		count := 1
		if strings.HasPrefix(t, "(") {
			re := regexp.MustCompile(`\((\d+)\)\s`)
			res := re.FindAllStringSubmatch(t, -1)
			if len(res) > 0 && len(res[0]) > 1 {
				count, _ = strconv.Atoi(res[0][1])
				count++
			}
		}
		status = "(" + strconv.Itoa(count) + ") " + status
	}
	lbl.SetText(status)
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
	b.logPopup.UI.Hide()
	b.outerInterface.FocusInterface()
	b.inPopup = false
}
