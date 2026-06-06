/******************************************************************************/
/* assertions.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package debug

import (
	"errors"
	"log/slog"
	"runtime"

	"kaijuengine.com/build"
)

var NotImplementedError = errors.New("not implemented")

func Log(msg string, args ...any) {
	if build.Debug {
		slog.Debug(msg, args...)
	}
}

func Assert(res bool, msg string) {
	if !res {
		if build.Debug {
			panic(msg)
		} else {
			slog.Error(msg)
		}
	}
}

func Halt(msg string) {
	if build.Debug {
		slog.Error(msg)
		runtime.Breakpoint()
	}
}

func Ensure(res bool) {
	if build.Debug {
		if !res {
			runtime.Breakpoint()
		}
	}
}

func EnsureMsg(res bool, msg string) {
	if build.Debug {
		if !res {
			slog.Error(msg)
			runtime.Breakpoint()
		}
	}
}

func EnsureNotError(err error) {
	if build.Debug {
		if err != nil {
			EnsureMsg(false, err.Error())
		}
	}
}

func EnsureNotNil(target any) {
	if build.Debug {
		EnsureMsg(target != nil, "the target was expected not to be nil, but was")
	}
}

func ThrowNotImplemented(todo string) { EnsureMsg(false, todo) }
func Throw(message string)            { Assert(false, message) }
