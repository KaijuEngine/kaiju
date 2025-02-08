/******************************************************************************/
/* log_window.go                                                              */
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

package log_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/logging"
	"kaiju/ui"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const sizeConfig = "logWindowSize"

type viewGroup = int

const (
	viewGroupAll viewGroup = iota
	viewGroupInfo
	viewGroupWarn
	viewGroupError
	viewGroupSelected
)

type visibleMessage struct {
	Time     string
	Message  string
	Trace    string
	Data     map[string]string
	Category string
}

func newVisibleMessage(msg string, trace []string, cat string) visibleMessage {
	mapping := logging.ToMap(msg)
	t, _ := time.Parse(time.RFC3339, mapping["time"])
	message := mapping["msg"]
	delete(mapping, "time")
	delete(mapping, "msg")
	return visibleMessage{
		Time:     t.Format(time.StampMilli),
		Message:  message,
		Trace:    strings.Join(trace, "\n"),
		Data:     mapping,
		Category: cat,
	}
}

type LogWindow struct {
	doc        *document.Document
	host       *engine.Host
	Group      viewGroup
	all        []visibleMessage
	lastReload engine.FrameId
	logStream  *logging.LogStream
	infoEvtId  logging.EventId
	warnEvtId  logging.EventId
	errEvtId   logging.EventId
	group      *ui.Group
	mutex      sync.Mutex
}

func New(host *engine.Host, logStream *logging.LogStream, uiGroup *ui.Group) *LogWindow {
	l := &LogWindow{
		lastReload: engine.InvalidFrameId,
		all:        make([]visibleMessage, 0),
		logStream:  logStream,
		host:       host,
		group:      uiGroup,
	}
	l.infoEvtId = logStream.OnInfo.Add(func(msg string) {
		l.add(msg, nil, "info")
	})
	l.warnEvtId = logStream.OnWarn.Add(func(msg string, trace []string) {
		l.add(msg, trace, "warn")
	})
	l.errEvtId = logStream.OnError.Add(func(msg string, trace []string) {
		l.add(msg, trace, "error")
	})
	host.OnClose.Add(func() {
		if l.doc != nil {
			l.doc.Destroy()
		}
	})
	return l
}

func (l *LogWindow) add(msg string, trace []string, cat string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.all = append(l.all, newVisibleMessage(msg, trace, cat))
	if l.isVisible() {
		l.reloadUI()
	}

}

func (l *LogWindow) All() []visibleMessage {
	res := slices.Clone(l.all)
	slices.Reverse(res)
	return res
}

func (l *LogWindow) filter(typeName string) []visibleMessage {
	res := make([]visibleMessage, 0, len(l.all))
	for i := range l.all {
		if l.all[i].Category == typeName {
			res = append(res, l.all[i])
		}
	}
	return res
}

func (l *LogWindow) Infos() []visibleMessage {
	res := l.filter("info")
	return res
}

func (l *LogWindow) Warnings() []visibleMessage {
	res := l.filter("warn")
	return res
}

func (l *LogWindow) Errors() []visibleMessage {
	res := l.filter("error")
	return res
}

func (l *LogWindow) isVisible() bool {
	return l.doc != nil && l.doc.Elements[0].UI.Entity().IsActive()
}

func (l *LogWindow) Toggle() {
	if l.doc == nil {
		l.Show()
	} else {
		if l.doc.Elements[0].UI.Entity().IsActive() {
			l.Hide()
		} else {
			l.Show()
		}
	}
}

func (l *LogWindow) Show() {
	if l.doc == nil {
		l.reloadUI()
	} else {
		l.doc.Activate()
		l.showCurrent()
	}
}

func (l *LogWindow) Hide() {
	if l.doc != nil {
		l.doc.Deactivate()
	}
}

func (l *LogWindow) clearAll(e *document.Element) {
	l.all = l.all[:0]
	l.reloadUI()
}

func (l *LogWindow) deactivateGroups() {
	all, _ := l.doc.GetElementById("all")
	info, _ := l.doc.GetElementById("info")
	warn, _ := l.doc.GetElementById("warn")
	err, _ := l.doc.GetElementById("error")
	selected, _ := l.doc.GetElementById("selected")
	all.UI.Entity().Deactivate()
	info.UI.Entity().Deactivate()
	warn.UI.Entity().Deactivate()
	err.UI.Entity().Deactivate()
	selected.UI.Entity().Deactivate()
	ab, _ := l.doc.GetElementById("allBtn")
	ib, _ := l.doc.GetElementById("infoBtn")
	wb, _ := l.doc.GetElementById("warningsBtn")
	eb, _ := l.doc.GetElementById("errorsBtn")
	sb, _ := l.doc.GetElementById("selectedBtn")
	ab.Children[0].UI.ToLabel().SetFontWeight("normal")
	ib.Children[0].UI.ToLabel().SetFontWeight("normal")
	wb.Children[0].UI.ToLabel().SetFontWeight("normal")
	eb.Children[0].UI.ToLabel().SetFontWeight("normal")
	sb.Children[0].UI.ToLabel().SetFontWeight("normal")
}

func (l *LogWindow) showCurrent() {
	switch l.Group {
	case viewGroupAll:
		l.showAll(nil)
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

func (l *LogWindow) showAll(*document.Element) {
	l.Group = viewGroupAll
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("all")
	b, _ := l.doc.GetElementById("allBtn")
	e.UI.Entity().Activate()
	b.Children[0].UI.ToLabel().SetFontWeight("bolder")
}

func (l *LogWindow) showInfos(*document.Element) {
	l.Group = viewGroupInfo
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("info")
	b, _ := l.doc.GetElementById("infoBtn")
	e.UI.Entity().Activate()
	b.Children[0].UI.ToLabel().SetFontWeight("bolder")
}

func (l *LogWindow) showWarns(*document.Element) {
	l.Group = viewGroupWarn
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("warn")
	b, _ := l.doc.GetElementById("warningsBtn")
	e.UI.Entity().Activate()
	b.Children[0].UI.ToLabel().SetFontWeight("bolder")
}

func (l *LogWindow) showErrors(*document.Element) {
	l.Group = viewGroupError
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("error")
	b, _ := l.doc.GetElementById("errorsBtn")
	e.UI.Entity().Activate()
	b.Children[0].UI.ToLabel().SetFontWeight("bolder")
}

func (l *LogWindow) showSelected(*document.Element) {
	l.Group = viewGroupSelected
	l.deactivateGroups()
	e, _ := l.doc.GetElementById("selected")
	b, _ := l.doc.GetElementById("selectedBtn")
	e.UI.Entity().Activate()
	b.Children[0].UI.ToLabel().SetFontWeight("bolder")
}

func (l *LogWindow) selectEntry(e *document.Element) {
	if id, err := strconv.Atoi(e.Attribute("data-entry")); err == nil {
		var target []visibleMessage
		switch l.Group {
		case viewGroupAll:
			target = l.all
		case viewGroupInfo:
			target = l.filter("info")
		case viewGroupWarn:
			target = l.filter("warn")
		case viewGroupError:
			target = l.filter("error")
		}
		if id >= 0 && id < len(target) {
			// The lists are printed in reverse order, so we invert the index
			id = len(target) - id - 1
			selectedElm, _ := l.doc.GetElementById("selected")
			lbl := selectedElm.Children[0].UI.ToLabel()
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
	const html = "editor/ui/log_window.html"
	if l.doc != nil {
		l.doc.Destroy()
	}
	frame := l.host.Frame()
	if l.lastReload == frame {
		return
	}
	l.host.CreatingEditorEntities()
	l.lastReload = frame
	l.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(
		l.host, html, l, map[string]func(*document.Element){
			"clearAll":     l.clearAll,
			"showAll":      l.showAll,
			"showInfos":    l.showInfos,
			"showWarns":    l.showWarns,
			"showErrors":   l.showErrors,
			"showSelected": l.showSelected,
			"selectEntry":  l.selectEntry,
			"resizeHover":  l.resizeHover,
			"resizeExit":   l.resizeExit,
			"resizeStart":  l.resizeStart,
			"resizeStop":   l.resizeStop,
		}))
	l.doc.SetGroup(l.group)
	l.host.DoneCreatingEditorEntities()
	l.showCurrent()
	l.doc.Clean()
	if s, ok := editor_cache.EditorConfigValue(sizeConfig); ok {
		w, _ := l.doc.GetElementById("window")
		if f32, ok := s.(float32); ok {
			w.UIPanel.Base().Layout().ScaleHeight(matrix.Float(f32))
		} else if f64, ok := s.(float64); ok {
			w.UIPanel.Base().Layout().ScaleHeight(matrix.Float(f64))
		}
	}
}

func (l *LogWindow) resizeHover(e *document.Element) {
	l.host.Window.CursorSizeNS()
}

func (l *LogWindow) resizeExit(e *document.Element) {
	dd := l.host.Window.Mouse.DragData()
	if dd != l {
		l.host.Window.CursorStandard()
	}
}

func (l *LogWindow) resizeStart(e *document.Element) {
	l.host.Window.CursorSizeNS()
	l.host.Window.Mouse.SetDragData(l)
}

func (l *LogWindow) resizeStop(e *document.Element) {
	dd := l.host.Window.Mouse.DragData()
	if dd != l {
		return
	}
	l.host.Window.CursorStandard()
	w, _ := l.doc.GetElementById("window")
	s := w.UIPanel.Base().Layout().PixelSize().Height()
	editor_cache.SetEditorConfigValue(sizeConfig, s)
}

func (l *LogWindow) DragUpdate() {
	w, _ := l.doc.GetElementById("window")
	y := l.host.Window.Mouse.Position().Y() - 20
	h := l.host.Window.Height()
	if int(y) < h-100 {
		w.UIPanel.Base().Layout().ScaleHeight(y)
	}
}
