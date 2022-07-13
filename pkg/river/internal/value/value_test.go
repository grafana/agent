package value_test

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

func TestFunction(t *testing.T) {
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
