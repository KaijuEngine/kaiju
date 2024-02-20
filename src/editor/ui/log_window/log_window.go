/******************************************************************************/
/* log_window.go                                                              */
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

package log_window

import (
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/logging"
	"kaiju/ui"
	"slices"
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
	Data    map[string]string
}

func newVisibleMessage(msg string, trace []string) visibleMessage {
	mapping := logging.ToMap(msg)
	t, _ := time.Parse(time.RFC3339, mapping["time"])
	message := mapping["msg"]
	delete(mapping, "time")
	delete(mapping, "msg")
	return visibleMessage{
		Time:    t.Format(time.StampMilli),
		Message: message,
		Trace:   strings.Join(trace, "\n"),
		Data:    mapping,
	}
}

type LogWindow struct {
	doc        *document.Document
	container  *host_container.Container
	Group      viewGroup
	infos      []visibleMessage
	warnings   []visibleMessage
	errors     []visibleMessage
	lastReload engine.FrameId
	logStream  *logging.LogStream
	infoEvtId  logging.EventId
	warnEvtId  logging.EventId
	errEvtId   logging.EventId
}

func (l *LogWindow) Infos() []visibleMessage {
	res := slices.Clone(l.infos)
	slices.Reverse(res)
	return res
}

func (l *LogWindow) Warnings() []visibleMessage {
	res := slices.Clone(l.warnings)
	slices.Reverse(res)
	return res
}

func (l *LogWindow) Errors() []visibleMessage {
	res := slices.Clone(l.errors)
	slices.Reverse(res)
	return res
}

func New(logStream *logging.LogStream) *LogWindow {
	l := &LogWindow{
		lastReload: engine.InvalidFrameId,
		infos:      make([]visibleMessage, 0),
		warnings:   make([]visibleMessage, 0),
		errors:     make([]visibleMessage, 0),
		logStream:  logStream,
	}
	l.infoEvtId = logStream.OnInfo.Add(func(msg string) {
		l.infos = append(l.infos, newVisibleMessage(msg, []string{}))
		l.reloadUI()
	})
	l.warnEvtId = logStream.OnWarn.Add(func(msg string, trace []string) {
		l.warnings = append(l.warnings, newVisibleMessage(msg, trace))
		l.reloadUI()
	})
	l.errEvtId = logStream.OnError.Add(func(msg string, trace []string) {
		l.errors = append(l.errors, newVisibleMessage(msg, trace))
		l.reloadUI()
	})
	return l
}

func (l *LogWindow) Show() {
	if l.container != nil {
		return
	}
	l.container = host_container.New("Log Window", nil)
	go l.container.Run(engine.DefaultWindowWidth, engine.DefaultWindowWidth/3)
	<-l.container.PrepLock
	l.reloadUI()
	l.container.Host.OnClose.Add(func() {
		l.logStream.OnInfo.Remove(l.infoEvtId)
		l.logStream.OnWarn.Remove(l.warnEvtId)
		l.logStream.OnError.Remove(l.errEvtId)
		l.container = nil
		l.lastReload = engine.InvalidFrameId
	})
}

func (l *LogWindow) clearAll(e *document.DocElement) {
	l.infos = l.infos[:0]
	l.warnings = l.warnings[:0]
	l.errors = l.errors[:0]
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
			target = l.infos
		case viewGroupWarn:
			target = l.warnings
		case viewGroupError:
			target = l.errors
		}
		if id >= 0 && id < len(target) {
			// The lists are printed in reverse order, so we invert the index
			id = len(target) - id - 1
			selectedElm, _ := l.doc.GetElementById("selected")
			lbl := selectedElm.HTML.Children[0].DocumentElement.UI.(*ui.Label)
			sb := strings.Builder{}
			sb.WriteString(target[id].Time)
			sb.WriteRune('\n')
			sb.WriteString(target[id].Message)
			sb.WriteRune('\n')
			for k, v := range target[id].Data {
				sb.WriteString(k)
				sb.WriteRune('=')
				sb.WriteString(v)
				sb.WriteRune('\n')
			}
			sb.WriteString(target[id].Trace)
			lbl.SetText(sb.String())
			l.showSelected(e)
		}
	}
}

func (l *LogWindow) reloadUI() {
	if l.container == nil {
		return
	}
	for _, e := range l.container.Host.Entities() {
		e.Destroy()
	}
	frame := l.container.Host.Frame()
	if l.lastReload == frame {
		return
	}
	l.lastReload = frame
	html := klib.MustReturn(l.container.Host.AssetDatabase().ReadText("editor/ui/log_window.html"))
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
