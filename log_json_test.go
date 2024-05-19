package logjson

import (
	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLogJson_MarshalNil(t *testing.T) {
	require.Equal(t, `null`, marshalToStr(nil))
}

func TestLogJson_MarshalString(t *testing.T) {
	require.Equal(t, `"hello"`, marshalToStr("hello"))
}

func TestLogJson_MarshalStruct(t *testing.T) {
	type Abc struct {
		Name string
	}
	require.Equal(t, `{"Name":"hello"}`, marshalToStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalPointer(t *testing.T) {
	type Abc struct {
		Pointer *Abc
		Name    string
	}
	require.Equal(t, `{"Pointer":null,"Name":"hello"}`, marshalToStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalOmit(t *testing.T) {
	ch := make(chan int, 5)
	require.Equal(t, `null`, marshalToStr(ch))
}

func TestLogJson_MarshalOmitTag(t *testing.T) {
	type Abc struct {
		Name string `log:"omit"`
	}
	require.Equal(t, `{}`, marshalToStr(Abc{Name: "hello"}))
}

func Test_StdMarshal(t *testing.T) {
	type Abc struct {
		Pointer *Abc
		Name    string
	}
	in := &Abc{
		Name: "hello",
	}
	json.Marshal(in)
}

func marshalToStr(in any) string {
	return string(Marshal(in))
}
