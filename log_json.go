package logjson

import (
	"github.com/go-json-experiment/json/jsontext"
	"reflect"
	"strings"
	"sync"
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
	case reflect.Struct:
		return j.makeStructHandlerItem(t)
	case reflect.Pointer:
		return j.makePointerHandlerItem(t)
	}
	return &handlerItem{
		marshal: func(v reflect.Value, state *encoderState) {
			state.encoder.WriteToken(jsontext.Null)
		},
	}
}

func (j *LogJson) makePointerHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	var once sync.Once
	var valItem *handlerItem
	init := func() {
		valItem = j.getHandlerItem(t.Elem())
	}
	item.marshal = func(v reflect.Value, state *encoderState) {
		if v.IsNil() {
			state.encoder.WriteToken(jsontext.Null)
			return
		}
		once.Do(init)
		valItem.marshal(v.Elem(), state)
	}
	return item
}

func (j *LogJson) makeStructHandlerItem(t reflect.Type) *handlerItem {
	var fields []structField
	var once sync.Once
	item := &handlerItem{}
	init := func() {
		fields = j.parseStructFields(t)
	}
	item.marshal = func(v reflect.Value, state *encoderState) {
		once.Do(init)
		state.encoder.WriteToken(jsontext.ObjectStart)
		for _, field := range fields {
			state.encoder.WriteToken(jsontext.String(field.Name))
			field.handlerItem.marshal(v.FieldByIndex(field.Index), state)
		}
		state.encoder.WriteToken(jsontext.ObjectEnd)
	}
	return item
}

type structField struct {
	Index       []int
	Name        string
	handlerItem *handlerItem
}

func newStructField(j *LogJson, field reflect.StructField) (structField, bool) {
	f := structField{}
	if !f.init(j, field) {
		return structField{}, false
	}
	return f, true
}

func (f *structField) init(j *LogJson, field reflect.StructField) bool {
	tagStr := field.Tag.Get("log")
	switch tagStr {
	case "omit":
		return false
	}
	f.Name = field.Name
	f.initJsonTag(field)
	f.Index = field.Index
	f.handlerItem = j.getHandlerItem(field.Type)
	return true
}

func (f *structField) initJsonTag(field reflect.StructField) {
	parts := strings.Split(field.Tag.Get("json"), ",")
	for i, part := range parts {
		if i == 0 {
			if part != "" {
				f.Name = part
			}
			continue
		}
	}
}

func (j *LogJson) parseStructFields(t reflect.Type) []structField {
	fields := reflect.VisibleFields(t)
	var result []structField
	for _, field := range fields {
		newField, ok := newStructField(j, field)
		if !ok {
			continue
		}
		result = append(result, newField)
	}
	return result
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
