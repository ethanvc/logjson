package logjson

import (
	"github.com/go-json-experiment/json/jsontext"
	"io"
	"reflect"
)

type Encoder struct {
	encoder *jsontext.Encoder
	w       io.Writer
	visited map[valueId]struct{}
}

func NewEncoder(w io.Writer) *Encoder {
	encoder := &Encoder{
		w: w,
	}
	encoder.encoder = jsontext.NewEncoder(w)
	return encoder
}

func (encoder *Encoder) GetWriter() io.Writer {
	return encoder.w
}

func (encoder *Encoder) Reset(w io.Writer) {
	encoder.encoder.Reset(w)
	encoder.w = w
	encoder.visited = nil
}

func (encoder *Encoder) enterPointer(v reflect.Value) bool {
	key := valueId{v.Type(), v.UnsafePointer(), encoder.sliceLen(v)}
	if _, ok := encoder.visited[key]; ok {
		return false
	} else {
		if encoder.visited == nil {
			encoder.visited = make(map[valueId]struct{})
		}
		encoder.visited[key] = struct{}{}
		return true
	}
}

func (encoder *Encoder) sliceLen(v reflect.Value) int {
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

func (encoder *Encoder) leavePointer(v reflect.Value) {
	key := valueId{v.Type(), v.UnsafePointer(), encoder.sliceLen(v)}
	delete(encoder.visited, key)
}
