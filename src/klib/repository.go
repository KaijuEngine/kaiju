/******************************************************************************/
/* repository.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"
)

var complaints = make(map[int]bool)

func NotYetImplemented(issueId int) {
	const msgPrefix = "This code is not yet implemented. If you are interested in contributing to the project by implementing this function, please see https://github.com/KaijuEngine/kaiju/issues/"
	if _, ok := complaints[issueId]; ok {
		return
	}
	complaints[issueId] = true
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	slog.LogAttrs(context.Background(), slog.LevelWarn,
		msgPrefix+strconv.Itoa(issueId),
		slog.String("file", frame.File),
		slog.Int("line", frame.Line),
		slog.String("function", frame.Function),
		slog.Int("issueId", issueId))
}
