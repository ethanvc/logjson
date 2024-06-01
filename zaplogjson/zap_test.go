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
		Name     string
		BankCard string `log:"md5"`
	}
	abc := Abc{
		Name:     "hello",
		BankCard: "",
	}
	logger.Info("Test", zap.Any("test", abc))
	require.Equal(t, `{"level":"info","msg":"Test","test":{"Name":"hello","BankCard":"0;d41d8cd98f00b204e9800998ecf8427e"}}`+"\n", buf.String())
}

func newTestZapLogger() (*zap.Logger, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	syncer := zapcore.AddSync(buf)
	encoderConf := zap.NewProductionEncoderConfig()
	encoderConf.TimeKey = ""
	encoderConf.NewReflectedEncoder = NewReflectEncoderFactory(nil).NewReflectedEncoder
	encoder := zapcore.NewJSONEncoder(encoderConf)
	core := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)
	logger := zap.New(core)
	return logger, buf
}
