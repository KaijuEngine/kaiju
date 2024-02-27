/******************************************************************************/
/* editor_status_bar.go                                                       */
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

package status_bar

import (
	"kaiju/editor/ui/log_window"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/logging"
	"kaiju/ui"
	"regexp"
	"strconv"
	"strings"
)

type StatusBar struct {
	doc *document.Document
	msg *ui.Label
	log *ui.Label
}

func New(host *engine.Host, logWindow *log_window.LogWindow) *StatusBar {
	const html = "editor/ui/status.html"
	s := &StatusBar{}
	s.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(host, html, nil,
		map[string]func(*document.DocElement){
			"openLogWindow": func(*document.DocElement) { logWindow.Show() },
		}))
	m, _ := s.doc.GetElementById("msg")
	l, _ := s.doc.GetElementById("log")
	s.msg = m.HTML.Children[0].DocumentElement.UI.(*ui.Label)
	s.log = l.HTML.Children[0].DocumentElement.UI.(*ui.Label)
	host.LogStream.OnInfo.Add(func(msg string) {
		host.RunAfterFrames(1, func() { s.setLog(msg, matrix.ColorWhite()) })
	})
	host.LogStream.OnWarn.Add(func(msg string, _ []string) {
		host.RunAfterFrames(1, func() { s.setLog(msg, matrix.ColorYellow()) })
	})
	host.LogStream.OnError.Add(func(msg string, _ []string) {
		host.RunAfterFrames(1, func() { s.setLog(msg, matrix.ColorLightCoral()) })
	})
	return s
}

func (s *StatusBar) setLog(msg string, color matrix.Color) {
	s.log.SetColor(color)
	res := logging.ToMap(msg)
	if m, ok := res["msg"]; ok {
		s.log.SetText(m)
	} else {
		s.log.SetText(msg)
	}
}

func (s *StatusBar) SetMessage(status string) {
	t := s.msg.Text()
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
	s.log.Host().RunAfterFrames(1, func() { s.msg.SetText(status) })
}
