package logjson

import (
	"github.com/go-json-experiment/json/jsontext"
)

func Marshal(in any) []byte {
	return DefaultLogJson().Marshal(in)
}

type LogMarshaler interface {
	MarshalLogJSON(*jsontext.Encoder)
}
