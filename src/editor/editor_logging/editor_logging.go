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

func (l *Logging) Initialize(host *engine.Host, logStream *logging.LogStream) {
	wl := weak.Make(l)
	l.infoEvtId = logStream.OnInfo.Add(func(msg string) {
		ll := wl.Value()
		if ll != nil {
			ll.add(msg, nil, "info")
		}
	})
	l.warnEvtId = logStream.OnWarn.Add(func(msg string, trace []string) {
		ll := wl.Value()
		if ll != nil {
			ll.add(msg, trace, "warn")
		}
	})
	l.errEvtId = logStream.OnError.Add(func(msg string, trace []string) {
		ll := wl.Value()
		if ll != nil {
			ll.add(msg, trace, "error")
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

func (l *Logging) add(msg string, trace []string, cat string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	m := newVisibleMessage(msg, trace, cat)
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

func newVisibleMessage(msg string, trace []string, cat string) Message {
	mapping := logging.ToMap(msg)
	t, _ := time.Parse(time.RFC3339, mapping["time"])
	message := mapping["msg"]
	delete(mapping, "time")
	delete(mapping, "msg")
	return Message{
		Time:     t.Format(time.StampMilli),
		Message:  message,
		Trace:    strings.Join(trace, "\n"),
		Data:     mapping,
		Category: cat,
	}
}
