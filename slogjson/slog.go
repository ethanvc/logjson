package slogjson

import (
	"context"
	"io"
	"log/slog"
)

type Handler struct {
}

func NewHandler(conf *HandlerOption) *Handler {
	h := &Handler{}
	return h
}

type HandlerOption struct {
	Writer io.Writer
}

func (h *Handler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (h *Handler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return h
}
