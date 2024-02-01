package klib

import (
	"fmt"
	"runtime"
)

func NotYetImplemented(issueId int) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fmt.Printf("%s:%d %s is not yet implemented. If you are interested in contributing to the project by implementing this function, please see https://github.com/KaijuEngine/kaiju/issues/%d\n", frame.File, frame.Line, frame.Function, issueId)
}
