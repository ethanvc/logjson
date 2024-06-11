package slogjson

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func Test_Basic(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	h := NewHandler(&HandlerOption{
		Writer: buf,
	})
	l := slog.New(h)
	require.Equal(t, ``, buf.String())
	type Abc struct {
		Name string
	}
	abc := Abc{
		Name: "test",
	}
	l.Info("Test", slog.String("xx", "abc"),
		slog.Any("abc", abc))
	require.Regexp(t, `.*|Test|{"xx":"abc","abc":{"Name":"test"}}`, buf.String())
}
