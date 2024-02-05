package klib

import (
	"fmt"
	"runtime"
	"strings"
)

func TraceString(message string) string {
	sb := strings.Builder{}
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	sb.WriteString(message)
	sb.WriteRune('\n')
	frame, ok := frames.Next()
	for ok {
		sb.WriteString(fmt.Sprintf("\t%s:%d %s\n", frame.File, frame.Line, frame.Function))
		frame, ok = frames.Next()
	}
	return sb.String()
}

func Trace(message string) {
	fmt.Print(TraceString(message))
}
