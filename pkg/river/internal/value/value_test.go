package value_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

// TestInterfacePointerReceiver tests various cases where Go values which only
// implement an interface for a pointer receiver are retained correctly
// throughout values.
func TestInterfacePointerReceiver(t *testing.T) {
	tt := []struct {
		name         string
		encodeTarget any
		key          string

		expectBodyType value.Type
		expectBodyText string

		expectKeyExists bool
		expectKeyValue  value.Value
		expectKeyType   value.Type
	}{
		{
			name:           "Encode",
			encodeTarget:   &pointerMarshaler{},
			expectBodyType: value.TypeString,
			expectBodyText: "Hello, world!",
		},
		{
			name:            "Struct Encode data Key",
			encodeTarget:    &Body{},
			key:             "data",
			expectBodyType:  value.TypeObject,
			expectKeyExists: true,
			expectKeyValue:  value.String("Hello, world!"),
			expectKeyType:   value.TypeString,
		},
		{
			name:            "Struct Encode Missing Key",
			encodeTarget:    &Body{},
			key:             "missing",
			expectBodyType:  value.TypeObject,
			expectKeyExists: false,
			expectKeyValue:  value.Null,
			expectKeyType:   value.TypeNull,
		},
		{
			name:            "Map Encode data Key",
			encodeTarget:    map[string]string{"data": "Hello, world!"},
			key:             "data",
			expectBodyType:  value.TypeObject,
			expectKeyExists: true,
			expectKeyValue:  value.String("Hello, world!"),
			expectKeyType:   value.TypeString,
		},
		{
			name:            "Map Encode Missing Key",
			encodeTarget:    map[string]string{"data": "Hello, world!"},
			key:             "missing",
			expectBodyType:  value.TypeObject,
			expectKeyExists: false,
			expectKeyValue:  value.Null,
			expectKeyType:   value.TypeNull,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			bodyVal := value.Encode(tc.encodeTarget)
			require.Equal(t, tc.expectBodyType, bodyVal.Type())
			if tc.expectBodyText != "" {
				require.Equal(t, "Hello, world!", bodyVal.Text())
			}

			if tc.key != "" {
				val, ok := bodyVal.Key(tc.key)
				require.Equal(t, tc.expectKeyExists, ok)
				require.Equal(t, tc.expectKeyType, val.Type())
				switch val.Type() {
				case value.TypeString:
					require.Equal(t, tc.expectKeyValue.Text(), val.Text())
				case value.TypeNull:
					require.Equal(t, tc.expectKeyValue, val)
				default:
					require.Fail(t, "unexpected value type (this switch can be expanded)")
				}
			}
		})
	}
}

type pointerMarshaler struct{}

func (*pointerMarshaler) MarshalText() ([]byte, error) {
	return []byte("Hello, world!"), nil
}

type Body struct {
	Data pointerMarshaler `river:"data,attr"`
}

func TestValue_Call(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		add := func(a, b int) int { return a + b }
		addVal := value.Encode(add)

		res, err := addVal.Call(
			value.Int(15),
			value.Int(43),
		)
		require.NoError(t, err)
		require.Equal(t, int64(15+43), res.Int())
	})

	t.Run("fully variadic", func(t *testing.T) {
		add := func(nums ...int) int {
			var sum int
			for _, num := range nums {
				sum += num
			}
			return sum
		}
		addVal := value.Encode(add)

		t.Run("no args", func(t *testing.T) {
			res, err := addVal.Call()
			require.NoError(t, err)
			require.Equal(t, int64(0), res.Int())
		})

		t.Run("one arg", func(t *testing.T) {
			res, err := addVal.Call(value.Int(32))
			require.NoError(t, err)
			require.Equal(t, int64(32), res.Int())
		})

		t.Run("many args", func(t *testing.T) {
			res, err := addVal.Call(
				value.Int(32),
				value.Int(59),
				value.Int(12),
			)
			require.NoError(t, err)
			require.Equal(t, int64(32+59+12), res.Int())
		})
	})

	t.Run("partially variadic", func(t *testing.T) {
		add := func(firstNum int, nums ...int) int {
			sum := firstNum
			for _, num := range nums {
				sum += num
			}
			return sum
		}
		addVal := value.Encode(add)

		t.Run("no variadic args", func(t *testing.T) {
			res, err := addVal.Call(value.Int(52))
			require.NoError(t, err)
			require.Equal(t, int64(52), res.Int())
		})

		t.Run("one variadic arg", func(t *testing.T) {
			res, err := addVal.Call(value.Int(52), value.Int(32))
			require.NoError(t, err)
			require.Equal(t, int64(52+32), res.Int())
		})

		t.Run("many variadic args", func(t *testing.T) {
			res, err := addVal.Call(
				value.Int(32),
				value.Int(59),
				value.Int(12),
			)
			require.NoError(t, err)
			require.Equal(t, int64(32+59+12), res.Int())
		})
	})

	t.Run("returns error", func(t *testing.T) {
		failWhenTrue := func(val bool) (int, error) {
			if val {
				return 0, fmt.Errorf("function failed for a very good reason")
			}
			return 0, nil
		}
		funcVal := value.Encode(failWhenTrue)

		t.Run("no error", func(t *testing.T) {
			res, err := funcVal.Call(value.Bool(false))
			require.NoError(t, err)
			require.Equal(t, int64(0), res.Int())
		})

		t.Run("error", func(t *testing.T) {
			_, err := funcVal.Call(value.Bool(true))
			require.EqualError(t, err, "function failed for a very good reason")
		})
	})
}

func TestValue_Interface_In_Array(t *testing.T) {
	type Container struct {
		Field io.Closer `river:"field,attr"`
	}

	val := value.Encode(Container{Field: io.NopCloser(nil)})
	fieldVal, ok := val.Key("field")
	require.True(t, ok, "field not found in object")
	require.Equal(t, value.TypeCapsule, fieldVal.Type())

	arrVal := value.Array(fieldVal)
	require.Equal(t, value.TypeCapsule, arrVal.Index(0).Type())
}
