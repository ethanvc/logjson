package logjson

import (
	"bytes"
	"github.com/go-json-experiment/json/jsontext"
	"reflect"
)

type encoderState struct {
	encoder *jsontext.Encoder
	buf     bytes.Buffer
	visited map[valueId]struct{}
}

func newEncoderState() *encoderState {
	state := &encoderState{}
	state.encoder = jsontext.NewEncoder(&state.buf)
	return state
}

func (state *encoderState) enterPointer(v reflect.Value) bool {
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

func (state *encoderState) sliceLen(v reflect.Value) int {
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

func (state *encoderState) leavePointer(v reflect.Value) {
	key := valueId{v.Type(), v.UnsafePointer(), state.sliceLen(v)}
	delete(state.visited, key)
}
