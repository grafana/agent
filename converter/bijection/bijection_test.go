package bijection

import (
	"github.com/stretchr/testify/require"
	"testing"
)

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

	bi := StructBijection[RiverExample, YamlExample]{}
	BindField(&bi, PropertyNames{A: "TestRiver", B: "TestYAML"}, Copy[int]())
	BindField(&bi, PropertyNames{A: "UInt", B: "UIntValue"}, Copy[uint]())
	BindField(&bi, PropertyNames{A: "Str", B: "String"}, Copy[string]())

	from := RiverExample{
		TestRiver: 42,
		UInt:      123,
		Str:       "hello",
	}
	expectedTo := YamlExample{
		TestYAML:  42,
		UIntValue: 123,
		String:    "hello",
	}

	testTwoWayConversion(t, bi, from, expectedTo)
}

func TestNested(t *testing.T) {
	type RiverExample struct {
		TestRiver int32
		UInt      uint64
		Str       string
	}

	type YamlExample struct {
		TestYAML  int64
		UIntValue float64
		Bytes     []byte
	}

	bi := StructBijection[RiverExample, YamlExample]{}

	int32ToInt64 := FnBijection[int32, int64]{
		AtoB: func(a *int32, b *int64) error {
			*b = int64(*a)
			return nil
		},
		BtoA: func(b *int64, a *int32) error {
			*a = int32(*b)
			return nil
		},
	}

	uint64ToFloat64 := FnBijection[uint64, float64]{
		AtoB: func(a *uint64, b *float64) error {
			*b = float64(*a)
			return nil
		},
		BtoA: func(b *float64, a *uint64) error {
			*a = uint64(*b)
			return nil
		},
	}

	BindField[RiverExample, YamlExample, int32, int64](&bi, PropertyNames{A: "TestRiver", B: "TestYAML"}, int32ToInt64)
	BindField[RiverExample, YamlExample, uint64, float64](&bi, PropertyNames{A: "UInt", B: "UIntValue"}, uint64ToFloat64)
	BindField(&bi, PropertyNames{A: "Str", B: "Bytes"}, Copy[string]())

	from := RiverExample{
		TestRiver: 42,
		UInt:      123,
		Str:       "hello",
	}
	expectedTo := YamlExample{
		TestYAML:  42,
		UIntValue: 123,
		Bytes:     []byte("hello"),
	}

	testTwoWayConversion(t, bi, from, expectedTo)
}

func testTwoWayConversion[A any, B any](t *testing.T, bi StructBijection[A, B], from A, expectedTo B) {
	var to B
	err := bi.ConvertAToB(&from, &to)
	require.NoError(t, err)
	require.Equal(t, expectedTo, to)

	var reversed A
	err = bi.ConvertBToA(&to, &reversed)
	require.NoError(t, err)
	require.Equal(t, from, reversed)
}
