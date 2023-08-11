package bijection

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	type RiverExample struct {
		TestRiver int
		UInt      uint
		Str       string
	}

	type YamlExample struct {
		TestYAML  int
		String    string
		UIntValue uint
	}

	bi := StructBijection[RiverExample, YamlExample]{
		Mappings: map[PropertyPair]interface{}{
			{A: "TestRiver", B: "TestYAML"}: Identiy,
			{A: "UInt", B: "UIntValue"}:     Identiy,
			{A: "Str", B: "String"}:         Identiy,
		},
	}

	from := RiverExample{
		TestRiver: 42,
		UInt:      123,
		Str:       "hello",
	}
	to := YamlExample{}

	err := bi.ConvertAToB(&from, &to)
	require.NoError(t, err)
	require.Equal(t, YamlExample{
		TestYAML:  42,
		UIntValue: 123,
		String:    "hello",
	}, to)

	reversed := RiverExample{}
	err = bi.ConvertBToA(&to, &reversed)
	require.NoError(t, err)
	require.Equal(t, from, reversed)

}

func TestSimple(t *testing.T) {
	type RiverExample struct {
		TestRiver int
		UInt      uint
		Str       string
	}

	type YamlExample struct {
		TestYAML  int
		String    string
		UIntValue uint
	}

	bi := NewBijection(RiverExample{}, YamlExample{})

	bi.Bind("TestRiver", "TestYAML", identity, identity)
	bi.Bind("UInt", "UIntValue", identity, identity)
	bi.Bind("Str", "String", identity, identity)

	from := RiverExample{
		TestRiver: 42,
		UInt:      123,
		Str:       "hello",
	}
	to := YamlExample{}

	err := bi.ConvertAToB(&from, &to)
	require.NoError(t, err)
	require.Equal(t, YamlExample{
		TestYAML:  42,
		UIntValue: 123,
		String:    "hello",
	}, to)

	reversed := RiverExample{}
	err = bi.ConvertBToA(&to, &reversed)
	require.NoError(t, err)
	require.Equal(t, from, reversed)
}

func TestConvertTypes(t *testing.T) {
	type RiverExample struct {
		TestRiver int
		UInt      uint
	}

	type YamlExample struct {
		TestYAML  int32
		UIntValue float32
	}

	a := RiverExample{}
	b := YamlExample{}
	bi := NewBijection(a, b)

	bi.Bind(
		"TestRiver",
		"TestYAML",
		func(a any) (any, error) {
			return int32(a.(int)), nil
		},
		func(a any) (any, error) {
			return int(a.(int32)), nil
		},
	)
	bi.Bind(
		"UInt",
		"UIntValue",
		func(a any) (any, error) {
			return float32(a.(uint)), nil
		},
		func(a any) (any, error) {
			return uint(a.(float32)), nil
		},
	)

	from := RiverExample{
		TestRiver: 42,
		UInt:      123,
	}
	to := YamlExample{}

	err := bi.ConvertAToB(&from, &to)
	require.NoError(t, err)
	require.Equal(t, YamlExample{
		TestYAML:  42,
		UIntValue: 123,
	}, to)

	reversed := RiverExample{}
	err = bi.ConvertBToA(&to, &reversed)
	require.NoError(t, err)
	require.Equal(t, from, reversed)
}
