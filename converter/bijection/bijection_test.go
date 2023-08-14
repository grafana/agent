package bijection

import (
	"testing"

	"github.com/stretchr/testify/require"
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

	sb := StructBijection[RiverExample, YamlExample]{}
	BindField(&sb, Names{A: "TestRiver", B: "TestYAML"}, Copy[int]())
	BindField(&sb, Names{A: "UInt", B: "UIntValue"}, Copy[uint]())
	BindField(&sb, Names{A: "Str", B: "String"}, Copy[string]())

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

	var b Bijection[RiverExample, YamlExample] = &sb
	testTwoWayConversion(t, b, from, expectedTo)
}

func TestCustomConversions(t *testing.T) {
	type RiverExample struct {
		TestRiver int64
		UInt      uint64
		Str       string
		Bytes     []byte
	}

	type YamlExample struct {
		TestYAML  int32
		UIntValue float64
		Bytes     []byte
		Str       string
	}

	sb := StructBijection[RiverExample, YamlExample]{}

	// Test inverting a bijection too
	inverted := Inverted(Cast[int32, int64]())
	BindField(&sb, Names{A: "TestRiver", B: "TestYAML"}, inverted)
	BindField(&sb, Names{A: "UInt", B: "UIntValue"}, Cast[uint64, float64]())
	BindField(&sb, Names{A: "Str", B: "Bytes"}, Cast[string, []byte]())
	BindField(&sb, Names{A: "Bytes", B: "Str"}, Cast[[]byte, string]())

	from := RiverExample{
		TestRiver: 42,
		UInt:      123,
		Str:       "hello",
		Bytes:     []byte("hello2"),
	}
	expectedTo := YamlExample{
		TestYAML:  42,
		UIntValue: 123,
		Bytes:     []byte("hello"),
		Str:       "hello2",
	}

	var b Bijection[RiverExample, YamlExample] = &sb
	testTwoWayConversion(t, b, from, expectedTo)
}

func testTwoWayConversion[A any, B any](t *testing.T, bi Bijection[A, B], from A, expectedTo B) {
	var to B
	err := bi.ConvertAToB(&from, &to)
	require.NoError(t, err)
	require.Equal(t, expectedTo, to)

	var reversed A
	err = bi.ConvertBToA(&to, &reversed)
	require.NoError(t, err)
	require.Equal(t, from, reversed)
}
