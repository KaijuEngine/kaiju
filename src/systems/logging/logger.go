package logging

import (
	"kaiju/klib"
	"log/slog"
	"os"
	"strings"
)

type LogStream struct {
	OnInfo  Event
	OnWarn  TracedEvent
	OnError TracedEvent
}

func (l *LogStream) Write(p []byte) (n int, err error) {
	str := string(p)
	levelOffset := 21
	if strings.HasPrefix(str, "time=") {
		levelOffset = 41
	}
	level := str[levelOffset:]
	if strings.HasPrefix(level, "INFO") {
		l.OnInfo.Execute(str)
	} else if strings.HasPrefix(level, "WARN") {
		l.OnWarn.Execute(str, klib.TraceStrings(str, 7))
	} else if strings.HasPrefix(level, "ERROR") {
		l.OnError.Execute(str, klib.TraceStrings(str, 7))
	}
	os.Stdout.WriteString(str)
	return len(p), nil
}

func Initialize(opts *slog.HandlerOptions) *LogStream {
	stream := &LogStream{
		OnInfo:  newEvent(),
		OnWarn:  newTracedEvent(),
		OnError: newTracedEvent(),
	}
	logger := slog.New(newLogHandler(stream, opts))
	slog.SetDefault(logger)
	return stream
}
