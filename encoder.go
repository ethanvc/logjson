package logjson

import (
	"bytes"
	"github.com/go-json-experiment/json/jsontext"
)

type encoderState struct {
	encoder *jsontext.Encoder
	buf     bytes.Buffer
}

func newEncoderState() *encoderState {
	state := &encoderState{}
	state.encoder = jsontext.NewEncoder(&state.buf)
	return state
}
