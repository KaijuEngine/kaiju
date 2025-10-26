/******************************************************************************/
/* editor_logging.go                                                          */
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

package editor_logging

import (
	"kaiju/engine"
	"kaiju/engine/systems/logging"
	"strings"
	"sync"
	"time"
	"weak"
)

type Logging struct {
	all       []Message
	infoEvtId logging.EventId
	warnEvtId logging.EventId
	errEvtId  logging.EventId
	OnNewLog  func(msg Message)
	mutex     sync.Mutex
}

type Message struct {
	Time     string
	Message  string
	Trace    string
	Data     map[string]string
	Category string
}

func (m *Message) ToString() string {
	endLen := len(m.Time) + 1 + len(m.Message)
	for k, v := range m.Data {
		endLen += 1 + len(k) + 1 + len(v)
	}
	sb := strings.Builder{}
	sb.Grow(endLen)
	sb.WriteString(m.Time)
	sb.WriteRune(' ')
	sb.WriteString(m.Message)
	for k, v := range m.Data {
		sb.WriteRune(' ')
		sb.WriteString(k)
		sb.WriteRune('=')
		sb.WriteString(v)
	}
	return sb.String()
}

func (l *Logging) Initialize(host *engine.Host, logStream *logging.LogStream) {
	wl := weak.Make(l)
	l.infoEvtId = logStream.OnInfo.Add(func(msg string) {
		ll := wl.Value()
		if ll != nil {
			ll.add(msg, nil)
		}
	})
	l.warnEvtId = logStream.OnWarn.Add(func(msg string, trace []string) {
		ll := wl.Value()
		if ll != nil {
			ll.add(msg, trace)
		}
	})
	l.errEvtId = logStream.OnError.Add(func(msg string, trace []string) {
		ll := wl.Value()
		if ll != nil {
			ll.add(msg, trace)
		}
	})
	host.OnClose.Add(func() {
		logStream.OnInfo.Remove(l.infoEvtId)
		logStream.OnWarn.Remove(l.warnEvtId)
		logStream.OnError.Remove(l.errEvtId)
	})
}

func (l *Logging) Clear()              { l.all = l.all[:0] }
func (l *Logging) All() []Message      { return l.all }
func (l *Logging) Infos() []Message    { return l.filter("info") }
func (l *Logging) Warnings() []Message { return l.filter("warn") }
func (l *Logging) Errors() []Message   { return l.filter("error") }

func (l *Logging) add(msg string, trace []string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	m := newVisibleMessage(msg, trace)
	l.all = append(l.all, m)
	if l.OnNewLog != nil {
		l.OnNewLog(m)
	}
}

func (l *Logging) filter(typeName string) []Message {
	res := make([]Message, 0, len(l.all))
	for i := range l.all {
		if l.all[i].Category == typeName {
			res = append(res, l.all[i])
		}
	}
	return res
}

func newVisibleMessage(msg string, trace []string) Message {
	mapping := logging.ToMap(msg)
	t, _ := time.Parse(time.RFC3339, mapping["time"])
	message := mapping["msg"]
	cat := strings.ToLower(mapping["level"])
	delete(mapping, "time")
	delete(mapping, "msg")
	delete(mapping, "level")
	return Message{
		Time:     t.Format(time.StampMilli),
		Message:  message,
		Trace:    strings.Join(trace, "\n"),
		Data:     mapping,
		Category: cat,
	}
}
