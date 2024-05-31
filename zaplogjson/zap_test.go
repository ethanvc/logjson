package zaplogjson

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestNewLogJsonEncoder(t *testing.T) {
	logger, buf := newTestZapLogger()
	type Abc struct {
		Name string
	}
	abc := Abc{
		Name: "hello",
	}
	logger.Info("Test", zap.Any("test", abc))
	require.Equal(t, `{"level":"info","msg":"Test","test":{"Name":"hello"}}`+"\n", buf.String())
}

func newTestZapLogger() (*zap.Logger, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	syncer := zapcore.AddSync(buf)
	encoderConf := zap.NewProductionEncoderConfig()
	encoderConf.TimeKey = ""
	encoder := NewJsonLogEncoder(NewJsonLogEncoderConfig(encoderConf))
	core := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)
	logger := zap.New(core)
	return logger, buf
}
