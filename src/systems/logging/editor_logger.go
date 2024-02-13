//go:build editor

package logging

import (
	"context"
	"io"
	"log/slog"
)

type EditorLogHandler struct {
	slog.Handler
}

func (e *EditorLogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func newLogHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return &EditorLogHandler{
		Handler: slog.NewTextHandler(w, opts),
	}
}
