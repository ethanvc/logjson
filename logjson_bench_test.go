package logjson

import (
	"github.com/go-json-experiment/json"
	"testing"
)

func Benchmark_LogJson(b *testing.B) {
	type Abc struct {
		Name string
	}
	abc := Abc{
		Name: "hello",
	}
	b.Run("LogJson", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Marshal(abc)
		}
	})
	b.Run("StdJson", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(abc)
		}
	})
}
