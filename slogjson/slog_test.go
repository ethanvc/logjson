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
	l.Info("Test")
	require.Equal(t, `test`, buf.String())
}
