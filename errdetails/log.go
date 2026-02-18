package errdetails

import (
	"context"
	"log/slog"
)

// NewNoopSlogLogger returns a *slog.Logger that discards all log output.
func NewNoopSlogLogger() *slog.Logger {
	return slog.New(noopHandler{})
}

// noopHandler is a slog.Handler that discards all log records.
type noopHandler struct{}

func (noopHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (noopHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (h noopHandler) WithAttrs(_ []slog.Attr) slog.Handler        { return h }
func (h noopHandler) WithGroup(_ string) slog.Handler              { return h }
