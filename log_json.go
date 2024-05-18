package logjson

import (
	"github.com/go-json-experiment/json/jsontext"
	"reflect"
)

type LogJson struct {
}

var defaultJson = NewLogJson()

func NewLogJson() *LogJson {
	return &LogJson{}
}

func (j *LogJson) Marshal(in any) []byte {
	state := newEncoderState()
	v := reflect.ValueOf(in)
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		state.encoder.WriteToken(jsontext.Null)
		return removeNewline(state.buf.Bytes())
	}
	j.getHandlerItem(v.Type()).marshal(v, state)
	return removeNewline(state.buf.Bytes())
}

func (j *LogJson) getHandlerItem(t reflect.Type) *handlerItem {
	switch t.Kind() {
	case reflect.String:
		return j.makeStringHandlerItem()
	}
	return nil
}

func (j *LogJson) makeStringHandlerItem() *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *encoderState) {
		state.encoder.WriteToken(jsontext.String(v.String()))
	}
	return item
}

type handlerItem struct {
	marshal func(v reflect.Value, state *encoderState)
}

func removeNewline(s []byte) []byte {
	l := len(s)
	if l == 0 {
		return s
	}
	if s[l-1] == '\n' {
		return s[:l-1]
	}
	return s
}
