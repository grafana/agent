package value_test

import (
	"reflect"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

var typeTests = []struct {
	input  interface{}
	expect value.Type
}{
	{int(0), value.TypeNumber},
	{int8(0), value.TypeNumber},
	{int16(0), value.TypeNumber},
	{int32(0), value.TypeNumber},
	{int64(0), value.TypeNumber},
	{uint(0), value.TypeNumber},
	{uint8(0), value.TypeNumber},
	{uint16(0), value.TypeNumber},
	{uint32(0), value.TypeNumber},
	{uint64(0), value.TypeNumber},
	{float32(0), value.TypeNumber},
	{float64(0), value.TypeNumber},

	{string(""), value.TypeString},

	{bool(false), value.TypeBool},

	{[...]int{0, 1, 2}, value.TypeArray},
	{[]int{0, 1, 2}, value.TypeArray},

	{struct{}{}, value.TypeObject},

	{map[string]interface{}{}, value.TypeObject},

	// Go functions must have one non-error return type and one optional error
	// return type to be River functions. Everything else is a capsule.
	{(func() int)(nil), value.TypeFunction},
	{(func() (int, error))(nil), value.TypeFunction},
	{(func())(nil), value.TypeCapsule},                 // Must have non-error return type
	{(func() error)(nil), value.TypeCapsule},           // First return type must be non-error
	{(func() (error, int))(nil), value.TypeCapsule},    // First return type must be non-error
	{(func() (int, int))(nil), value.TypeCapsule},      // Second return type must be error
	{(func() (int, int, int))(nil), value.TypeCapsule}, // Can only have 1 or 2 return types

	{make(chan struct{}), value.TypeCapsule},
	{map[bool]interface{}{}, value.TypeCapsule}, // Maps with non-string types are capsules
}

func Test_RiverType(t *testing.T) {
	for _, tc := range typeTests {
		rt := reflect.TypeOf(tc.input)

		t.Run(rt.String(), func(t *testing.T) {
			actual := value.RiverType(rt)
			require.Equal(t, tc.expect, actual, "Unexpected type for %#v", tc.input)
		})
	}
}

func Benchmark_RiverType(b *testing.B) {
	for _, tc := range typeTests {
		rt := reflect.TypeOf(tc.input)

		b.Run(rt.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = value.RiverType(rt)
			}
		})
	}
}
