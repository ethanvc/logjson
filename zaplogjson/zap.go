package zaplogjson

import (
	"github.com/ethanvc/logjson"
	"go.uber.org/zap/zapcore"
	"io"
	"sync"
)

type ReflectEncoderFactory struct {
	j *logjson.LogJson
}

func NewReflectEncoderFactory(j *logjson.LogJson) *ReflectEncoderFactory {
	if j == nil {
		j = logjson.DefaultLogJson()
	}
	return &ReflectEncoderFactory{
		j: j,
	}
}

func (factory *ReflectEncoderFactory) NewReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	return ReflectedEncoder{
		w: w,
		j: factory.j,
	}
}

// ReflectedEncoder implement zapcore.ReflectedEncoder
type ReflectedEncoder struct {
	w io.Writer
	j *logjson.LogJson
}

func (enc ReflectedEncoder) Encode(val any) error {
	state, _ := GetFromSyncPool[*logjson.EncoderState](encoderStatePool)
	if state == nil {
		state = logjson.NewEncoderState(enc.w)
	} else {
		state.Reset(enc.w)
	}
	enc.j.MarshalWithState(val, state)
	encoderStatePool.Put(state)
	return nil
}

var encoderStatePool = &sync.Pool{}

func GetFromSyncPool[V any](p *sync.Pool) (V, bool) {
	v := p.Get()
	if vv, ok := v.(V); ok {
		return vv, true
	} else {
		var v V
		return v, false
	}
}
