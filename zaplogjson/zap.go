package zaplogjson

import (
	"github.com/ethanvc/logjson"
	"go.uber.org/zap/zapcore"
	"io"
)

func NewReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	return ReflectedEncoder{
		w: w,
	}
}

// ReflectedEncoder implement zapcore.ReflectedEncoder
type ReflectedEncoder struct {
	w io.Writer
}

func (enc ReflectedEncoder) Encode(val any) error {
	buf := logjson.Marshal(val)
	enc.w.Write(buf)
	return nil
}

type JsonLogEncoder struct {
	zapcore.Encoder
}

func NewJsonLogEncoder(conf JsonLogEncoderConfig) *JsonLogEncoder {
	encoder := &JsonLogEncoder{
		Encoder: zapcore.NewJSONEncoder(conf.EncoderConfig),
	}
	return encoder
}

type JsonLogEncoderConfig struct {
	zapcore.EncoderConfig
}

func NewJsonLogEncoderConfig(conf zapcore.EncoderConfig) JsonLogEncoderConfig {
	return JsonLogEncoderConfig{
		EncoderConfig: conf,
	}
}
