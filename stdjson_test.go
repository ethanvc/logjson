package logjson

import (
	jsonv1 "encoding/json"
	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

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

func Test_StdMarshal_MapKey(t *testing.T) {
	type Abc struct {
		Name string
	}
	m := map[Abc]int{
		Abc{"Hello"}: 3,
	}
	buf, err := json.Marshal(m)
	require.Error(t, err)
	_ = buf
	m1 := map[*int32]int{
		proto.Int32(3): 3,
	}
	var p *int32
	m1[p] = 4
	buf, err = json.Marshal(m1)
	require.Error(t, err)
}

func Test_StdMarshal_BoolKey(t *testing.T) {
	m := map[bool]int32{
		true: 3,
	}
	buf, err := json.Marshal(m)
	_ = buf
	require.Equal(t, `jsontext: missing string for object name`, err.Error())
	buf, err = jsonv1.Marshal(m)
	require.Equal(t, `json: unsupported type: map[bool]int32`, err.Error())
}
