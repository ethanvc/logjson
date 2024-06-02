package slogjson

import (
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func Test_Basic(t *testing.T) {
	l := slog.New(nil)
	require.NotNil(t, l)
}
