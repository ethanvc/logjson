package logjson

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetFilePathForLog(t *testing.T) {
	require.Equal(t, "logjson/stack_test.go:0",
		GetFilePathForLog("abc/logjson/stack_test.go", 0))
	require.Equal(t, ":0",
		GetFilePathForLog("", 0))
}
