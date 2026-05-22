/******************************************************************************/
/* trace.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"fmt"
	"runtime"
	"strings"
)

func TraceStrings(message string, skip int) []string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	result := make([]string, 0)
	frame, ok := frames.Next()
	for ok {
		result = append(result, fmt.Sprintf("\t%s:%d %s", frame.File, frame.Line, frame.Function))
		frame, ok = frames.Next()
	}
	return result
}

func TraceString(message string) string {
	return strings.Join(TraceStrings(message, 3), "\n")
}

func Trace(message string) {
	fmt.Print(TraceString(message))
}
