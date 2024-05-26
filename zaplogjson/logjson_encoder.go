package zaplogjson

import "go.uber.org/zap/zapcore"

func NewLogJsonEncoder(conf zapcore.EncoderConfig) *LogJsonEncoder {
	encoder := &LogJsonEncoder{}
	encoder.init(conf)
	return encoder
}

type LogJsonEncoder struct {
	zapcore.Encoder
}

func (encoder *LogJsonEncoder) init(conf zapcore.EncoderConfig) {
	encoder.Encoder = zapcore.NewJSONEncoder(conf)
}
