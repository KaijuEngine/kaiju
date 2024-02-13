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
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/systems/logging"
	"log/slog"
	"strings"
)

type viewGroup = int

const (
	viewGroupInfo viewGroup = iota
	viewGroupWarn
	viewGroupError
)

type visibleMessage struct {
	Time    string
	Message string
	Trace   string
}

func newVisibleMessage(msg string, trace []string) visibleMessage {
	msg = strings.TrimPrefix(msg, "time=")
	timeEnd := strings.Index(msg, " ")
	msgStart := strings.Index(msg, "msg=")
	return visibleMessage{
		Time:    msg[:timeEnd],
		Message: msg[msgStart+4:],
		Trace:   strings.Join(trace, "\n"),
	}
}

type LogWindow struct {
	doc       *document.Document
	container *host_container.Container
	Group     viewGroup
	Infos     []visibleMessage
	Warnings  []visibleMessage
	Errors    []visibleMessage
}

func (l *LogWindow) ShowingInfos() bool  { return l.Group == viewGroupInfo }
func (l *LogWindow) ShowingWarns() bool  { return l.Group == viewGroupWarn }
func (l *LogWindow) ShowingErrors() bool { return l.Group == viewGroupError }

func New(logStream *logging.LogStream) *LogWindow {
	l := &LogWindow{
		container: host_container.New("Log Window", nil),
	}
	go l.container.Run(500, 600)
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

var count = 0

func (l *LogWindow) clearAll(e *document.DocElement) {
	count++
	slog.Info("Clearing all logs", slog.Int("Count", count))
	slog.Warn("Clearing all logs", slog.Int("Count", count))
	slog.Error("Clearing all logs", slog.Int("Count", count))
	//l.Infos = l.Infos[:0]
	//l.Warnings = l.Warnings[:0]
	//l.Errors = l.Errors[:0]
	//l.reloadUI()
}

func (l *LogWindow) showInfos(e *document.DocElement) {
	l.Group = viewGroupInfo
	infoElm, _ := l.doc.GetElementById("info")
	warnElm, _ := l.doc.GetElementById("warn")
	errorElm, _ := l.doc.GetElementById("error")
	infoElm.UI.Entity().Activate()
	warnElm.UI.Entity().Deactivate()
	errorElm.UI.Entity().Deactivate()
}

func (l *LogWindow) showWarns(e *document.DocElement) {
	l.Group = viewGroupWarn
	infoElm, _ := l.doc.GetElementById("info")
	warnElm, _ := l.doc.GetElementById("warn")
	errorElm, _ := l.doc.GetElementById("error")
	infoElm.UI.Entity().Deactivate()
	warnElm.UI.Entity().Activate()
	errorElm.UI.Entity().Deactivate()
}

func (l *LogWindow) showErrors(e *document.DocElement) {
	l.Group = viewGroupError
	infoElm, _ := l.doc.GetElementById("info")
	warnElm, _ := l.doc.GetElementById("warn")
	errorElm, _ := l.doc.GetElementById("error")
	infoElm.UI.Entity().Deactivate()
	warnElm.UI.Entity().Deactivate()
	errorElm.UI.Entity().Activate()
}

func (l *LogWindow) selectEntry(e *document.DocElement) {
	slog.Info("Selected entry", slog.String("Name", e.HTML.Attribute("data-entry")))
}

func (l *LogWindow) reloadUI() {
	for _, e := range l.container.Host.Entities() {
		e.Destroy()
	}
	html := klib.MustReturn(l.container.Host.AssetDatabase().ReadText("ui/editor/log_window.html"))
	l.doc = markup.DocumentFromHTMLString(l.container.Host, html, "", l, map[string]func(*document.DocElement){
		"clearAll":    l.clearAll,
		"showInfos":   l.showInfos,
		"showWarns":   l.showWarns,
		"showErrors":  l.showErrors,
		"selectEntry": l.selectEntry,
	})
}
