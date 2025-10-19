package editor_logging

import (
	"kaiju/engine"
	"kaiju/engine/systems/logging"
	"strings"
	"sync"
	"time"
)

type Logging struct {
	all       []visibleMessage
	infoEvtId logging.EventId
	warnEvtId logging.EventId
	errEvtId  logging.EventId
	mutex     sync.Mutex
}

type visibleMessage struct {
	Time     string
	Message  string
	Trace    string
	Data     map[string]string
	Category string
}

func (l *Logging) Initialize(host *engine.Host, logStream *logging.LogStream) {
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
		logStream.OnInfo.Remove(l.infoEvtId)
		logStream.OnWarn.Remove(l.warnEvtId)
		logStream.OnError.Remove(l.errEvtId)
	})
}

func (l *Logging) Clear()                     { l.all = l.all[:0] }
func (l *Logging) All() []visibleMessage      { return l.all }
func (l *Logging) Infos() []visibleMessage    { return l.filter("info") }
func (l *Logging) Warnings() []visibleMessage { return l.filter("warn") }
func (l *Logging) Errors() []visibleMessage   { return l.filter("error") }

func (l *Logging) add(msg string, trace []string, cat string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.all = append(l.all, newVisibleMessage(msg, trace, cat))
}

func (l *Logging) filter(typeName string) []visibleMessage {
	res := make([]visibleMessage, 0, len(l.all))
	for i := range l.all {
		if l.all[i].Category == typeName {
			res = append(res, l.all[i])
		}
	}
	return res
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
