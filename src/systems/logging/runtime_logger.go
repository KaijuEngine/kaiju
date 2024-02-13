//go:build !editor

package logging

import (
	"context"
	"io"
	"log/slog"
)

type RuntimeLogHandler struct {
	slog.Handler
}

func (e *RuntimeLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelWarn
}

func newLogHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return &RuntimeLogHandler{
		Handler: slog.NewTextHandler(w, opts),
	}
}
