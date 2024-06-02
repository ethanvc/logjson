package logjson

import (
	"github.com/go-json-experiment/json/jsontext"
	"io"
	"reflect"
)

type EncoderState struct {
	*jsontext.Encoder
	w       io.Writer
	visited map[valueId]struct{}
}

func NewEncoderState(w io.Writer) *EncoderState {
	encoder := &EncoderState{
		w: w,
	}
	encoder.Encoder = jsontext.NewEncoder(w)
	return encoder
}

func (state *EncoderState) GetWriter() io.Writer {
	return state.w
}

func (state *EncoderState) Reset(w io.Writer) {
	state.Encoder.Reset(w)
	state.w = w
	state.visited = nil
}

func (state *EncoderState) enterPointer(v reflect.Value) bool {
	key := valueId{v.Type(), v.UnsafePointer(), state.sliceLen(v)}
	if _, ok := state.visited[key]; ok {
		return false
	} else {
		if state.visited == nil {
			state.visited = make(map[valueId]struct{})
		}
		state.visited[key] = struct{}{}
		return true
	}
}

func (state *EncoderState) sliceLen(v reflect.Value) int {
	if v.Kind() == reflect.Slice {
		return v.Len()
	}
	return 0
}

type valueId struct {
	t reflect.Type
	p any
	l int
}

func (state *EncoderState) leavePointer(v reflect.Value) {
	key := valueId{v.Type(), v.UnsafePointer(), state.sliceLen(v)}
	delete(state.visited, key)
}
