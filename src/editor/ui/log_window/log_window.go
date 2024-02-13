/*****************************************************************************/
/* log_window.go                                                             */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package log_window

import (
	"fmt"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/logging"
	"kaiju/ui"
	"strconv"
	"strings"
	"time"
)

type viewGroup = int

const (
	viewGroupInfo viewGroup = iota
	viewGroupWarn
	viewGroupError
	viewGroupSelected
)

type visibleMessage struct {
	Time    string
	Message string
	Trace   string
}

func newVisibleMessage(msg string, trace []string) visibleMessage {
	mapping := logging.ToMap(msg)
	t, _ := time.Parse(time.RFC3339, mapping["time"])
	return visibleMessage{
		Time:    t.Format(time.StampMilli),
		Message: mapping["msg"],
		Trace:   strings.Join(trace, "\n"),
	}
}

type LogWindow struct {
	doc        *document.Document
	container  *host_container.Container
	Group      viewGroup
	Infos      []visibleMessage
	Warnings   []visibleMessage
	Errors     []visibleMessage
	lastReload float64
}

func New(logStream *logging.LogStream) *LogWindow {
	l := &LogWindow{
		container:  host_container.New("Log Window", nil),
		lastReload: -1,
		Infos:      make([]visibleMessage, 0),
		Warnings:   make([]visibleMessage, 0),
		Errors:     make([]visibleMessage, 0),
	}
	go l.container.Run(engine.DefaultWindowWidth, engine.DefaultWindowWidth/3)
	<-l.container.PrepLock
	l.reloadUI()
	iID := logStream.OnInfo.Add(func(msg string) {
		l.Infos = append(l.Infos, newVisibleMessage(msg, []string{}))
		l.reloadUI()
	})
	wID := logStream.OnWarn.Add(func(msg string, trace []string) {
		l.Warnings = append(l.Warnings, newVisibleMessage(msg, trace))
		l.reloadUI()
	})
	eID := logStream.OnError.Add(func(msg string, trace []string) {
		l.Errors = append(l.Errors, newVisibleMessage(msg, trace))
		l.reloadUI()
	})
	l.container.Host.OnClose.Add(func() {
		logStream.OnInfo.Remove(iID)
		logStream.OnWarn.Remove(wID)
		logStream.OnError.Remove(eID)
	})
	return l
}

func (l *LogWindow) clearAll(e *document.DocElement) {
	l.Infos = l.Infos[:0]
	l.Warnings = l.Warnings[:0]
	l.Errors = l.Errors[:0]
	l.reloadUI()
}

func (l *LogWindow) deactivateGroups() {
	info, _ := l.doc.GetElementById("info")
	warn, _ := l.doc.GetElementById("warn")
	err, _ := l.doc.GetElementById("error")
	selected, _ := l.doc.GetElementById("selected")
	info.UI.Entity().Deactivate()
	warn.UI.Entity().Deactivate()
	err.UI.Entity().Deactivate()
	selected.UI.Entity().Deactivate()
	ib, _ := l.doc.GetElementById("infoBtn")
	wb, _ := l.doc.GetElementById("warningsBtn")
	eb, _ := l.doc.GetElementById("errorsBtn")
	sb, _ := l.doc.GetElementById("selectedBtn")
	ib.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("normal")
	wb.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("normal")
	eb.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("normal")
	sb.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("normal")
}

func (l *LogWindow) showCurrent() {
	switch l.Group {
	case viewGroupInfo:
		l.showInfos(nil)
	case viewGroupWarn:
		l.showWarns(nil)
	case viewGroupError:
		l.showErrors(nil)
	case viewGroupSelected:
		l.showSelected(nil)
	}
}

func (l *LogWindow) showInfos(*document.DocElement) {
	l.Group = viewGroupInfo
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("info")
	b, _ := l.doc.GetElementById("infoBtn")
	e.UI.Entity().Activate()
	b.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("bolder")
}

func (l *LogWindow) showWarns(*document.DocElement) {
	l.Group = viewGroupWarn
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("warn")
	b, _ := l.doc.GetElementById("warningsBtn")
	e.UI.Entity().Activate()
	b.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("bolder")
}

func (l *LogWindow) showErrors(*document.DocElement) {
	l.Group = viewGroupError
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("error")
	b, _ := l.doc.GetElementById("errorsBtn")
	e.UI.Entity().Activate()
	b.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("bolder")
}

func (l *LogWindow) showSelected(*document.DocElement) {
	l.Group = viewGroupSelected
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("selected")
	b, _ := l.doc.GetElementById("selectedBtn")
	e.UI.Entity().Activate()
	b.HTML.Children[0].DocumentElement.UI.(*ui.Label).SetFontWeight("bolder")
}

func (l *LogWindow) selectEntry(e *document.DocElement) {
	if id, err := strconv.Atoi(e.HTML.Attribute("data-entry")); err == nil {
		var target []visibleMessage
		switch l.Group {
		case viewGroupInfo:
			target = l.Infos
		case viewGroupWarn:
			target = l.Warnings
		case viewGroupError:
			target = l.Errors
		}
		if id >= 0 && id < len(target) {
			selectedElm, _ := l.doc.GetElementById("selected")
			lbl := selectedElm.HTML.Children[0].DocumentElement.UI.(*ui.Label)
			lbl.SetText(fmt.Sprintf("%s\n%s\n\n%s", target[id].Time,
				target[id].Message, target[id].Trace))
			l.showSelected(e)
		}
	}
}

func (l *LogWindow) reloadUI() {
	for _, e := range l.container.Host.Entities() {
		e.Destroy()
	}
	rt := l.container.Host.Runtime()
	if l.lastReload == rt {
		return
	}
	l.lastReload = rt
	html := klib.MustReturn(l.container.Host.AssetDatabase().ReadText("ui/editor/log_window.html"))
	l.container.RunFunction(func() {
		l.doc = markup.DocumentFromHTMLString(l.container.Host, html, "", l, map[string]func(*document.DocElement){
			"clearAll":     l.clearAll,
			"showInfos":    l.showInfos,
			"showWarns":    l.showWarns,
			"showErrors":   l.showErrors,
			"showSelected": l.showSelected,
			"selectEntry":  l.selectEntry,
		})
		l.showCurrent()
	})
}
