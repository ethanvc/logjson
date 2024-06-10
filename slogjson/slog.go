package slogjson

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethanvc/logjson"
	"github.com/go-json-experiment/json/jsontext"
	"io"
	"log/slog"
	"time"
)

type Handler struct {
	w io.Writer
	l *logjson.LogJson
}

func NewHandler(conf *HandlerOption) *Handler {
	conf.init()
	h := &Handler{
		w: conf.Writer,
		l: conf.LogJson,
	}
	return h
}

type HandlerOption struct {
	Writer  io.Writer
	LogJson *logjson.LogJson
}

func (o *HandlerOption) init() {
	if o.LogJson == nil {
		o.LogJson = logjson.DefaultLogJson()
	}
}

func (h *Handler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *Handler) Handle(c context.Context, record slog.Record) error {
	buf := bytes.NewBuffer(nil)
	state := logjson.NewEncoderState(buf)
	state.WriteToken(jsontext.ObjectStart)
	h.appendNonBuiltIns(state, record)
	state.WriteToken(jsontext.ObjectEnd)
	h.appendNewLineIfNeed(buf)
	h.w.Write(buf.Bytes())
	return nil
}

func (h *Handler) appendNewLineIfNeed(buf *bytes.Buffer) {
	by := buf.Bytes()
	if len(by) > 0 && by[len(by)-1] != '\n' {
		buf.WriteByte('\n')
	}
}

func (h *Handler) appendNonBuiltIns(state *logjson.EncoderState, r slog.Record) {
	r.Attrs(func(a slog.Attr) bool {
		h.appendItem(state, a)
		return true
	})
}

func (h *Handler) appendItem(state *logjson.EncoderState, a slog.Attr) {
	switch a.Value.Kind() {
	case slog.KindString:
		state.WriteToken(jsontext.String(a.Key))
		state.WriteToken(jsontext.String(a.Value.String()))
	case slog.KindUint64:
		state.WriteToken(jsontext.String(a.Key))
		state.WriteToken(jsontext.Uint(a.Value.Uint64()))
	case slog.KindInt64:
		state.WriteToken(jsontext.String(a.Key))
		state.WriteToken(jsontext.Int(a.Value.Int64()))
	case slog.KindFloat64:
		state.WriteToken(jsontext.String(a.Key))
		state.WriteToken(jsontext.Float(a.Value.Float64()))
	case slog.KindBool:
		state.WriteToken(jsontext.String(a.Key))
		state.WriteToken(jsontext.Bool(a.Value.Bool()))
	case slog.KindDuration:
		state.WriteToken(jsontext.String(a.Key))
		s := fmt.Sprintf("%dus", a.Value.Duration().Microseconds())
		state.WriteToken(jsontext.String(s))
	case slog.KindTime:
		state.WriteToken(jsontext.String(a.Key))
		s := a.Value.Time().Format(time.RFC3339Nano)
		state.WriteToken(jsontext.String(s))
	case slog.KindAny:
		state.WriteToken(jsontext.String(a.Key))
		h.l.MarshalWithState(a.Value.Any(), state)
	default:

	}
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return h
}
