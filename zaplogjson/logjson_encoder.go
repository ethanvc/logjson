package zaplogjson

import "go.uber.org/zap/zapcore"

func NewLogJsonEncoder(conf zapcore.EncoderConfig) *LogJsonEncoder {
	encoder := &LogJsonEncoder{}
	return encoder
}

type LogJsonEncoder struct {
	zapcore.Encoder
}

func (encoder *LogJsonEncoder) init(conf zapcore.EncoderConfig) {
	encoder.Encoder = zapcore.NewJSONEncoder(conf)
}

func (encoder *LogJsonEncoder) AddReflected(key string, value any) error {
	return nil
}
