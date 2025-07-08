package debug

import (
	"kaiju/build"
	"log/slog"
	"runtime"
)

func Log(msg string, args ...any) {
	if build.Debug {
		slog.Debug(msg, args...)
	}
}

func Assert(res bool, msg string) {
	if build.Debug {
		if !res {
			panic(msg)
		}
	} else {
		slog.Error(msg)
	}
}

func Ensure(res bool, msg string) {
	if !build.Shipping {
		if !res {
			slog.Error(msg)
			runtime.Breakpoint()
		}
	}
}
