package hcltypes

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func TestOptionalSecret(t *testing.T) {
	t.Run("non-sensitive conversion to string is allowed", func(t *testing.T) {
		input := OptionalSecret{Sensitive: false, Value: "testval"}

		result, err := convert.Convert(cty.CapsuleVal(optionalSecretTy, &input), cty.String)
		require.NoError(t, err)
		require.Equal(t, input.Value, result.AsString())
	})

	t.Run("sensitive conversion to string is disallowed", func(t *testing.T) {
		input := OptionalSecret{Sensitive: true, Value: "testval"}

		result, err := convert.Convert(cty.CapsuleVal(optionalSecretTy, &input), cty.String)
		require.EqualError(t, err, "cannot convert secret to string")
		require.Equal(t, cty.NilVal, result)
	})

	t.Run("non-sensitive conversion to secret is allowed", func(t *testing.T) {
		input := OptionalSecret{Sensitive: false, Value: "secretval"}

		result, err := convert.Convert(cty.CapsuleVal(optionalSecretTy, &input), secretTy)
		require.NoError(t, err)
		require.Equal(t, input.Value, string(*result.EncapsulatedValue().(*Secret)))
	})

	t.Run("sensitive conversion to secret is allowed", func(t *testing.T) {
		input := OptionalSecret{Sensitive: true, Value: "secretval"}

		result, err := convert.Convert(cty.CapsuleVal(optionalSecretTy, &input), secretTy)
		require.NoError(t, err)
		require.Equal(t, input.Value, string(*result.EncapsulatedValue().(*Secret)))
	})

	t.Run("conversion from string is allowed", func(t *testing.T) {
		input := "hello, world!"

		result, err := convert.Convert(cty.StringVal(input), optionalSecretTy)
		require.NoError(t, err)

		os := result.EncapsulatedValue().(*OptionalSecret)
		require.False(t, os.Sensitive)
		require.Equal(t, os.Value, input)
	})

	t.Run("conversion from secret is allowed", func(t *testing.T) {
		input := Secret("sensitive")

		result, err := convert.Convert(cty.CapsuleVal(secretTy, &input), optionalSecretTy)
		require.NoError(t, err)

		os := result.EncapsulatedValue().(*OptionalSecret)
		require.True(t, os.Sensitive)
		require.Equal(t, os.Value, string(input))
	})
}

func TestOptionalSecret_Write(t *testing.T) {
	type testBlock struct {
		Value OptionalSecret `hcl:"value,attr"`
	}

	t.Run("non-sensitive", func(t *testing.T) {
		b := testBlock{
			Value: OptionalSecret{Sensitive: false, Value: "not-hidden"},
		}

		f := hclwrite.NewFile()
		gohcl.EncodeIntoBody(&b, f.Body())
		require.Equal(t, "value = \"not-hidden\"\n", string(f.Bytes()))
	})

	t.Run("sensitive", func(t *testing.T) {
		b := testBlock{
			Value: OptionalSecret{Sensitive: true, Value: "hidden"},
		}

		f := hclwrite.NewFile()
		gohcl.EncodeIntoBody(&b, f.Body())
		require.Equal(t, "value = (secret)\n", string(f.Bytes()))
	})
}
