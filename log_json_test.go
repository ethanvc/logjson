package logjson

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLogJson_MarshalNil(t *testing.T) {
	require.Equal(t, `null`, marshalToStr(nil))
}

func TestLogJson_MarshalString(t *testing.T) {
	require.Equal(t, `"hello"`, marshalToStr("hello"))
}

func marshalToStr(in any) string {
	return string(Marshal(in))
}
