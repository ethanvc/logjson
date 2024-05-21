package logjson

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestLogJson_MarshalNil(t *testing.T) {
	require.Equal(t, `null`, marshalToLogStr(nil))
}

func TestLogJson_MarshalString(t *testing.T) {
	require.Equal(t, `"hello"`, marshalToLogStr("hello"))
}

func TestLogJson_MarshalStruct(t *testing.T) {
	type Abc struct {
		Name string
	}
	require.Equal(t, `{"Name":"hello"}`, marshalToLogStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalPointer(t *testing.T) {
	type Abc struct {
		Pointer *Abc
		Name    string
	}
	require.Equal(t, `{"Pointer":null,"Name":"hello"}`, marshalToLogStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalOmit(t *testing.T) {
	ch := make(chan int, 5)
	require.Equal(t, `null`, marshalToLogStr(ch))
}

func TestLogJson_MarshalOmitTag(t *testing.T) {
	type Abc struct {
		Name string `log:"omit"`
	}
	require.Equal(t, `{}`, marshalToLogStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalMd5Tag(t *testing.T) {
	type Abc struct {
		Name string `log:"md5"`
	}
	require.Equal(t, `{"Name":"5;5d41402abc4b2a76b9719d911017c592"}`,
		marshalToLogStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalWithJsonTag(t *testing.T) {
	type Abc struct {
		Name string `json:"name"`
	}
	require.Equal(t, `{"name":"hello"}`, marshalToLogStr(Abc{Name: "hello"}))
}

func TestLogJson_MarshalEmbedStruct(t *testing.T) {
	type Abc struct {
		Name string
	}
	type Bcd struct {
		Abc
	}
	require.Equal(t, `{"Name":"hello"}`,
		marshalToLogStr(Bcd{Abc: Abc{Name: "hello"}}))
}

func TestLogJson_MarshalUnexportedField(t *testing.T) {
	type Abc struct {
		name string
	}
	require.Equal(t, `{"name":"hello"}`,
		marshalToLogStr(Abc{name: "hello"}))
}

func TestLogJson_MarshalOmitEmpty(t *testing.T) {
	type Abc struct {
		name string `json:",omitempty"`
	}
	require.Equal(t, `{}`,
		marshalToLogStr(Abc{name: ""}))
}

func TestLogJson_Cycle(t *testing.T) {
	type Abc struct {
		p *Abc
	}
	abc := &Abc{}
	abc.p = abc
	require.Contains(t, marshalToLogStr(abc), `{"p":{"p":null}`)
}

func TestLogJson_Map(t *testing.T) {
	m := map[int]int{
		3: 3,
	}
	require.Equal(t, `{"3":3}`, marshalToLogStr(m))
}

func TestLogJson_Bytes(t *testing.T) {
	buf := []byte("hello")
	require.Equal(t, `"aGVsbG8"`, marshalToLogStr(buf))
}

func TestLogJson_Array(t *testing.T) {
	a := []int{3, 4, 5}
	require.Equal(t, `[3,4,5]`, marshalToLogStr(a))
}

func TestLogJson_Interface(t *testing.T) {
	num := 3
	var v any
	v = num
	require.Equal(t, `3`, marshalToLogStr(v))
}

func Test_ReflectString(t *testing.T) {
	num := 3
	v := reflect.ValueOf(num)
	require.Equal(t, "<int Value>", v.String())
}

func marshalToLogStr(in any) string {
	return string(Marshal(in))
}
