package logjson

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-json-experiment/json/jsontext"
	"reflect"
	"strconv"
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

var errorIntType = reflect.TypeFor[error]()

func (j *LogJson) getHandlerItem(t reflect.Type) *handlerItem {
	if t.Implements(errorIntType) {
		return j.makeErrorHandlerItem()
	}
	switch t.Kind() {
	case reflect.Bool:
		return j.makeBoolHandlerItem()
	case reflect.String:
		return j.makeStringHandlerItem()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return j.makeIntHandlerItem(t)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return j.makeUintHandlerItem(t)
	case reflect.Float32, reflect.Float64:
		return j.makeDoubleHandlerItem()
	case reflect.Map:
		return j.makeMapHandlerItem(t)
	case reflect.Struct:
		return j.makeStructHandlerItem(t)
	case reflect.Slice:
		return j.makeSliceHandlerItem(t)
	case reflect.Array:
		return j.makeArrayHandlerItem(t)
	case reflect.Pointer:
		return j.makePointerHandlerItem(t)
	case reflect.Interface:
		return j.makeInterfaceHandlerItem(t)
	}
	return &handlerItem{
		marshal: func(v reflect.Value, state *encoderState) {
			state.encoder.WriteToken(jsontext.Null)
		},
	}
}

func (j *LogJson) makeErrorHandlerItem() *handlerItem {
	return &handlerItem{
		marshal: func(v reflect.Value, state *encoderState) {
			if err, ok := v.Interface().(error); ok {
				state.encoder.WriteToken(jsontext.String(err.Error()))
				return
			} else {
				state.encoder.WriteToken(jsontext.Null)
			}
		},
	}
}

func (j *LogJson) makeInterfaceHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *encoderState) {
		if v.IsNil() {
			state.encoder.WriteToken(jsontext.Null)
			return
		}
		v = v.Elem()
		j.getHandlerItem(v.Type()).marshal(v, state)
	}
	return item
}

func (j *LogJson) makeArrayHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	var once sync.Once
	var elementHandlerItem *handlerItem
	init := func() {
		elementHandlerItem = j.getHandlerItem(t.Elem())
	}
	n := t.Len()
	item.marshal = func(v reflect.Value, state *encoderState) {
		once.Do(init)
		state.encoder.WriteToken(jsontext.ArrayStart)
		for i := 0; i < n; i++ {
			elementHandlerItem.marshal(v.Index(i), state)
		}
		state.encoder.WriteToken(jsontext.ArrayEnd)
	}
	return item
}

func (j *LogJson) makeMapHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	keyStringify, ok := generateMarshalToStringFunc(t.Key())
	if !ok {
		return &handlerItem{
			marshal: func(v reflect.Value, state *encoderState) {
				state.encoder.WriteToken(jsontext.Null)
			},
		}
	}
	var once sync.Once
	var valueHandlerItem *handlerItem
	init := func() {
		valueHandlerItem = j.getHandlerItem(t.Elem())
	}
	item.marshal = func(v reflect.Value, state *encoderState) {
		if v.IsNil() {
			state.encoder.WriteToken(jsontext.Null)
			return
		}
		if state.encoder.StackDepth() > startDetectingCyclesAfter {
			if !state.enterPointer(v) {
				state.encoder.WriteToken(jsontext.Null)
				return
			}
			defer state.leavePointer(v)
		}
		once.Do(init)
		state.encoder.WriteToken(jsontext.ObjectStart)
		for iter := v.MapRange(); iter.Next(); {
			tmp := keyStringify(iter.Key())
			state.encoder.WriteToken(jsontext.String(tmp))
			valueHandlerItem.marshal(iter.Value(), state)
		}
		state.encoder.WriteToken(jsontext.ObjectEnd)
	}
	return item
}

func (j *LogJson) makeBoolHandlerItem() *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *encoderState) {
		state.encoder.WriteToken(jsontext.Bool(v.Bool()))
	}
	return item
}

func (j *LogJson) makeDoubleHandlerItem() *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *encoderState) {
		state.encoder.WriteToken(jsontext.Float(v.Float()))
	}
	return item
}

func (j *LogJson) makeUintHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *encoderState) {
		state.encoder.WriteToken(jsontext.Uint(v.Uint()))
	}
	return item
}

func (j *LogJson) makeIntHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *encoderState) {
		state.encoder.WriteToken(jsontext.Int(v.Int()))
	}
	return item
}

func (j *LogJson) makeSliceHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	if t.Elem().Kind() == reflect.Uint8 {
		item.marshal = func(v reflect.Value, state *encoderState) {
			val := v.Bytes()
			base64Val := base64.RawStdEncoding.EncodeToString(val)
			state.encoder.WriteToken(jsontext.String(base64Val))
		}
		return item
	}
	var sliceItem *handlerItem
	var once sync.Once
	init := func() {
		sliceItem = j.getHandlerItem(t.Elem())
	}
	item.marshal = func(v reflect.Value, state *encoderState) {
		if v.IsNil() {
			state.encoder.WriteToken(jsontext.Null)
			return
		}
		if state.encoder.StackDepth() > startDetectingCyclesAfter {
			if !state.enterPointer(v) {
				state.encoder.WriteToken(jsontext.Null)
				return
			}
			defer state.leavePointer(v)
		}
		once.Do(init)
		n := v.Len()
		state.encoder.WriteToken(jsontext.ArrayStart)
		for i := 0; i < n; i++ {
			sliceItem.marshal(v.Index(i), state)
		}
		state.encoder.WriteToken(jsontext.ArrayEnd)
	}
	return item
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
		if state.encoder.StackDepth() > startDetectingCyclesAfter {
			if !state.enterPointer(v) {
				state.encoder.WriteToken(jsontext.Null)
				return
			}
			defer state.leavePointer(v)
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
			elmV := v.FieldByIndex(field.Index)
			if field.omitempty && isLegacyEmpty(elmV) {
				continue
			}
			state.encoder.WriteToken(jsontext.String(field.Name))
			field.handlerItem.marshal(elmV, state)
		}
		state.encoder.WriteToken(jsontext.ObjectEnd)
	}
	return item
}

type structField struct {
	Index       []int
	Name        string
	handlerItem *handlerItem
	omitempty   bool
}

func newStructField(j *LogJson, field reflect.StructField) (structField, bool) {
	f := structField{}
	if !f.init(j, field) {
		return structField{}, false
	}
	return f, true
}

func (f *structField) init(j *LogJson, field reflect.StructField) bool {
	f.Name = field.Name
	f.Index = field.Index
	if !f.initLogTag(j, field) {
		return false
	}
	f.initJsonTag(field)
	if f.handlerItem == nil {
		f.handlerItem = j.getHandlerItem(field.Type)
	}
	return true
}

func (f *structField) initLogTag(j *LogJson, field reflect.StructField) bool {
	switch field.Tag.Get("log") {
	case "omit":
		return false
	case "md5":
		if field.Type.Kind() == reflect.String {
			f.handlerItem = &handlerItem{}
			f.handlerItem.marshal = func(v reflect.Value, state *encoderState) {
				s := v.String()
				hexMd5 := md5.Sum([]byte(s))
				md5Str := hex.EncodeToString(hexMd5[:])
				state.encoder.WriteToken(jsontext.String(fmt.Sprintf("%d;%s", len(s), md5Str)))
			}
		}
	}
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
		switch part {
		case "omitempty":
			f.omitempty = true
		}
	}
}

func (j *LogJson) parseStructFields(t reflect.Type) []structField {
	fields := reflect.VisibleFields(t)
	var result []structField
	for _, field := range fields {
		if field.Anonymous {
			continue
		}
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

// isLegacyEmpty reports whether a value is empty according to the v1 definition.
func isLegacyEmpty(v reflect.Value) bool {
	// Equivalent to encoding/json.isEmptyValue@v1.21.0.
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func generateMarshalToStringFunc(t reflect.Type) (func(v reflect.Value) string, bool) {
	var cb func(v reflect.Value) string
	switch t.Kind() {
	case reflect.Bool:
		cb = func(v reflect.Value) string {
			return strconv.FormatBool(v.Bool())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cb = func(v reflect.Value) string {
			return strconv.FormatInt(v.Int(), 10)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cb = func(v reflect.Value) string {
			return strconv.FormatUint(v.Uint(), 10)
		}
	case reflect.Float32, reflect.Float64:
		cb = func(v reflect.Value) string {
			return strconv.FormatFloat(v.Float(), 'f', -1, 64)
		}
	case reflect.String:
		cb = func(v reflect.Value) string {
			return v.String()
		}
	}
	if cb != nil {
		return cb, true
	} else {
		return nil, false
	}
}

const startDetectingCyclesAfter = 1000
