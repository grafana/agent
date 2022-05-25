package hcltypes

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func TestSecret(t *testing.T) {
	t.Run("strings can be converted to secret", func(t *testing.T) {
		expect := "hello, world!"

		secretVal, err := convert.Convert(cty.StringVal(expect), secretTy)
		require.NoError(t, err)

		actual := secretVal.EncapsulatedValue().(*Secret)
		require.Equal(t, string(*actual), expect)
	})

	t.Run("secrets cannot be converted to strings", func(t *testing.T) {
		s := Secret("hello, world!")

		result, err := convert.Convert(cty.CapsuleVal(secretTy, &s), cty.String)
		require.EqualError(t, err, "string required")
		require.Equal(t, cty.NilVal, result)
	})

	t.Run("secrets can be passed to secrets", func(t *testing.T) {
		s := Secret("hello, world!")

		result, err := convert.Convert(cty.CapsuleVal(secretTy, &s), secretTy)
		require.NoError(t, err)

		actual := result.EncapsulatedValue().(*Secret)
		require.Equal(t, (*actual), s)
	})
}

func TestSecret_Write(t *testing.T) {
	type testBlock struct {
		Value Secret `hcl:"value,attr"`
	}

	b := testBlock{Value: Secret("sensitive")}

	f := hclwrite.NewFile()
	gohcl.EncodeIntoBody(&b, f.Body())
	require.Equal(t, "value = (secret)\n", string(f.Bytes()))
}
