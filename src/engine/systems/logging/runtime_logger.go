/******************************************************************************/
/* runtime_logger.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package logging

import (
	"context"
	"io"
	"log/slog"

	"kaijuengine.com/build"
)

type RuntimeLogHandler struct {
	slog.Handler
}

func (e *RuntimeLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	if build.Debug {
		return true
	} else {
		return level >= slog.LevelWarn
	}
}

func newLogHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if build.Editor {
		return &RuntimeLogHandler{
			Handler: slog.NewTextHandler(w, opts),
		}
	} else {
		return &RuntimeLogHandler{
			Handler: slog.NewJSONHandler(w, opts),
		}
	}
}
