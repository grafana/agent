package value_test

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/river/value"
)

func BenchmarkObjectDecode(b *testing.B) {
	b.StopTimer()

	// Create a value with 20 keys.
	source := make(map[string]string, 20)
	for i := 0; i < 20; i++ {
		var (
			key   = fmt.Sprintf("key_%d", i+1)
			value = fmt.Sprintf("value_%d", i+1)
		)
		source[key] = value
	}

	sourceVal := value.Encode(source)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		var dst map[string]string
		_ = value.Decode(sourceVal, &dst)
	}
}
