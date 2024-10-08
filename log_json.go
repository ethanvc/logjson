package logjson

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type LogJson struct {
	handlerItems sync.Map
	mux          sync.Mutex
	logRules     map[string]*logRuleConf
}

var defaultLogJson = NewLogJson()

func DefaultLogJson() *LogJson {
	return defaultLogJson
}

func NewLogJson() *LogJson {
	return &LogJson{
		logRules: make(map[string]*logRuleConf),
	}
}

func (j *LogJson) AddLogRule(key string, rule LogRule) {
	conf := newLogRuleConf(rule)
	j.mux.Lock()
	defer j.mux.Unlock()
	j.logRules[key] = conf
}

func (j *LogJson) Marshal(in any) []byte {
	var encoder *EncoderState
	var buf *bytes.Buffer
	encoderAny := encoderStatePool.Get()
	if encoderAny == nil {
		buf = bytes.NewBuffer(nil)
		encoder = NewEncoderState(buf)
	} else {
		encoder = encoderAny.(*EncoderState)
		buf = encoder.GetWriter().(*bytes.Buffer)
		buf.Reset()
		encoder.Reset(buf)
	}
	defer encoderStatePool.Put(encoder)
	j.MarshalWithState(in, encoder)
	return removeNewline(buf.Bytes())
}

func (j *LogJson) MarshalWithState(in any, encoder *EncoderState) {
	v := reflect.ValueOf(in)
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		encoder.Encoder.WriteToken(jsontext.Null)
		return
	}
	j.getHandlerItem(v.Type()).marshal(v, encoder)
}

func (j *LogJson) getLogRule(key string) *logRuleConf {
	j.mux.Lock()
	defer j.mux.Unlock()
	return j.logRules[key]
}

var errorIntType = reflect.TypeFor[error]()
var logMarshalerIntType = reflect.TypeFor[LogMarshaler]()
var marshalerV2IntType = reflect.TypeFor[json.MarshalerV2]()

func (j *LogJson) getHandlerItemInternal(t reflect.Type) *handlerItem {
	if t.Implements(logMarshalerIntType) {
		return j.makeLogMarshalerHandlerItem()
	}
	if t.Implements(marshalerV2IntType) {
		return j.makeMarshalerV2HandlerItem()
	}
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
		marshal: func(v reflect.Value, state *EncoderState) {
			state.Encoder.WriteToken(jsontext.Null)
		},
	}
}

func (j *LogJson) makeMarshalerV2HandlerItem() *handlerItem {
	return &handlerItem{
		marshal: func(v reflect.Value, state *EncoderState) {
			realInt, _ := v.Interface().(json.MarshalerV2)
			realInt.MarshalJSONV2(state.Encoder, nil)
		},
	}
}

func (j *LogJson) makeLogMarshalerHandlerItem() *handlerItem {
	return &handlerItem{
		marshal: func(v reflect.Value, state *EncoderState) {
			realInt, _ := v.Interface().(LogMarshaler)
			realInt.MarshalLogJSON(state.Encoder)
		},
	}
}

func (j *LogJson) getHandlerItem(t reflect.Type) *handlerItem {
	if tmp, ok := j.handlerItems.Load(t); ok {
		return tmp.(*handlerItem)
	}
	handler := j.getHandlerItemInternal(t)
	if existHandler, loaded := j.handlerItems.LoadOrStore(t, handler); loaded {
		return existHandler.(*handlerItem)
	}
	return handler
}

func (j *LogJson) makeErrorHandlerItem() *handlerItem {
	return &handlerItem{
		marshal: func(v reflect.Value, state *EncoderState) {
			if err, ok := v.Interface().(error); ok {
				state.Encoder.WriteToken(jsontext.String(err.Error()))
				return
			} else {
				state.Encoder.WriteToken(jsontext.Null)
			}
		},
	}
}

func (j *LogJson) makeInterfaceHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		if v.IsNil() {
			state.Encoder.WriteToken(jsontext.Null)
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
	item.marshal = func(v reflect.Value, state *EncoderState) {
		once.Do(init)
		state.Encoder.WriteToken(jsontext.ArrayStart)
		for i := 0; i < n; i++ {
			elementHandlerItem.marshal(v.Index(i), state)
		}
		state.Encoder.WriteToken(jsontext.ArrayEnd)
	}
	return item
}

func (j *LogJson) makeMapHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	keyStringify, ok := generateMarshalToStringFunc(t.Key())
	if !ok {
		return &handlerItem{
			marshal: func(v reflect.Value, state *EncoderState) {
				state.Encoder.WriteToken(jsontext.Null)
			},
		}
	}
	var once sync.Once
	var valueHandlerItem *handlerItem
	init := func() {
		valueHandlerItem = j.getHandlerItem(t.Elem())
	}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		if v.IsNil() {
			state.Encoder.WriteToken(jsontext.Null)
			return
		}
		if state.Encoder.StackDepth() > startDetectingCyclesAfter {
			if !state.enterPointer(v) {
				state.Encoder.WriteToken(jsontext.Null)
				return
			}
			defer state.leavePointer(v)
		}
		once.Do(init)
		state.Encoder.WriteToken(jsontext.ObjectStart)
		for iter := v.MapRange(); iter.Next(); {
			tmp := keyStringify(iter.Key())
			state.Encoder.WriteToken(jsontext.String(tmp))
			valueHandlerItem.marshal(iter.Value(), state)
		}
		state.Encoder.WriteToken(jsontext.ObjectEnd)
	}
	return item
}

func (j *LogJson) makeBoolHandlerItem() *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		state.Encoder.WriteToken(jsontext.Bool(v.Bool()))
	}
	return item
}

func (j *LogJson) makeDoubleHandlerItem() *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		state.Encoder.WriteToken(jsontext.Float(v.Float()))
	}
	return item
}

func (j *LogJson) makeUintHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		state.Encoder.WriteToken(jsontext.Uint(v.Uint()))
	}
	return item
}

func (j *LogJson) makeIntHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		state.Encoder.WriteToken(jsontext.Int(v.Int()))
	}
	return item
}

func (j *LogJson) makeSliceHandlerItem(t reflect.Type) *handlerItem {
	item := &handlerItem{}
	if t.Elem().Kind() == reflect.Uint8 {
		item.marshal = func(v reflect.Value, state *EncoderState) {
			val := v.Bytes()
			base64Val := base64.RawStdEncoding.EncodeToString(val)
			state.Encoder.WriteToken(jsontext.String(base64Val))
		}
		return item
	}
	var sliceItem *handlerItem
	var once sync.Once
	init := func() {
		sliceItem = j.getHandlerItem(t.Elem())
	}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		if v.IsNil() {
			state.Encoder.WriteToken(jsontext.Null)
			return
		}
		if state.Encoder.StackDepth() > startDetectingCyclesAfter {
			if !state.enterPointer(v) {
				state.Encoder.WriteToken(jsontext.Null)
				return
			}
			defer state.leavePointer(v)
		}
		once.Do(init)
		n := v.Len()
		state.Encoder.WriteToken(jsontext.ArrayStart)
		for i := 0; i < n; i++ {
			sliceItem.marshal(v.Index(i), state)
		}
		state.Encoder.WriteToken(jsontext.ArrayEnd)
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
	item.marshal = func(v reflect.Value, state *EncoderState) {
		if v.IsNil() {
			state.Encoder.WriteToken(jsontext.Null)
			return
		}
		if state.Encoder.StackDepth() > startDetectingCyclesAfter {
			if !state.enterPointer(v) {
				state.Encoder.WriteToken(jsontext.Null)
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
	item.marshal = func(v reflect.Value, state *EncoderState) {
		once.Do(init)
		state.Encoder.WriteToken(jsontext.ObjectStart)
		for _, field := range fields {
			elmV := v.FieldByIndex(field.Index)
			if field.omitempty && isLegacyEmpty(elmV) {
				continue
			}
			state.Encoder.WriteToken(jsontext.String(field.Name))
			field.handlerItem.marshal(elmV, state)
		}
		state.Encoder.WriteToken(jsontext.ObjectEnd)
	}
	return item
}

type structField struct {
	Index       []int
	Name        string
	handlerItem *handlerItem
	omitempty   bool
	omit        bool
	conf        *logRuleConf
}

func newStructField(j *LogJson, parentType reflect.Type, field reflect.StructField) structField {
	f := structField{}
	f.init(j, parentType, field)
	if f.conf != nil {
		if f.Omit() {
			return f
		}
		f.handlerItem = f.conf.GetHandlerItem(field)
	}
	if f.handlerItem == nil {
		f.handlerItem = j.getHandlerItem(field.Type)
	}
	return f
}

func (f *structField) init(j *LogJson, parentType reflect.Type, field reflect.StructField) {
	f.handlerItem = j.getHandlerItem(field.Type)
	f.Name = field.Name
	f.Index = field.Index
	f.initJsonTag(field)
	f.conf = newLogRuleConfFromStr(field.Tag.Get("log"))
	if f.conf != nil {
		return
	}
	protoLogJsonStr := getFieldOptionFromType(parentType, f.Name)
	f.conf = newLogRuleConfFromStr(protoLogJsonStr)
	if f.conf != nil {
		return
	}
	f.conf = j.getLogRule(f.Name)
	if f.conf != nil {
		return
	}
	return
}

func (f *structField) Omit() bool {
	if f.omit {
		return true
	}
	if f.conf != nil && f.conf.Omit() {
		return true
	}
	return false
}

func createMd5Marshal(t reflect.Type) func(v reflect.Value, state *EncoderState) {
	if t.Kind() == reflect.String {
		return func(v reflect.Value, state *EncoderState) {
			s := v.String()
			hexMd5 := md5.Sum([]byte(s))
			md5Str := hex.EncodeToString(hexMd5[:])
			state.Encoder.WriteToken(jsontext.String(fmt.Sprintf("%d;%s", len(s), md5Str)))
		}
	}
	if t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.String {
		return func(v reflect.Value, state *EncoderState) {
			if v.IsNil() {
				state.Encoder.WriteToken(jsontext.Null)
				return
			}
			s := v.Elem().String()
			hexMd5 := md5.Sum([]byte(s))
			md5Str := hex.EncodeToString(hexMd5[:])
			state.Encoder.WriteToken(jsontext.String(fmt.Sprintf("%d;%s", len(s), md5Str)))
		}
	}
	return nil
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
		if !field.IsExported() {
			continue
		}
		newField := newStructField(j, t, field)
		if newField.Omit() {
			continue
		}
		result = append(result, newField)
	}
	return result
}

func (j *LogJson) makeStringHandlerItem() *handlerItem {
	item := &handlerItem{}
	item.marshal = func(v reflect.Value, state *EncoderState) {
		state.Encoder.WriteToken(jsontext.String(v.String()))
	}
	return item
}

type handlerItem struct {
	marshal func(v reflect.Value, state *EncoderState)
}

func removeNewline(s []byte) []byte {
	l := len(s)
	if l == 0 {
		return nil
	}
	if s[l-1] == '\n' {
		return bytes.Clone(s[:l-1])
	}
	return bytes.Clone(s)
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

// reference: https://stackoverflow.com/questions/69311708/how-to-retrieve-fieldoption-value
func getFieldOptionLogJsonValue(v reflect.Value, key string) string {
	msg, ok := v.Interface().(proto.Message)
	if !ok {
		return ""
	}
	pfMsg := msg.ProtoReflect()
	msgDesc := pfMsg.Descriptor()
	fields := msgDesc.Fields()
	field := fields.ByName(protoreflect.Name(key))
	if field == nil {
		return ""
	}
	logJsonValAny := proto.GetExtension(field.Options(), E_LogJson)
	if s, ok := logJsonValAny.(string); ok {
		return s
	}
	return ""
}

func getFieldOptionFromType(t reflect.Type, key string) string {
	v := reflect.New(t)
	return getFieldOptionLogJsonValue(v, key)
}

const startDetectingCyclesAfter = 1000

var encoderStatePool sync.Pool
