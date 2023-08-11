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
	BindField(&bi, Names{A: "TestRiver", B: "TestYAML"}, Copy[int]())
	BindField(&bi, Names{A: "UInt", B: "UIntValue"}, Copy[uint]())
	BindField(&bi, Names{A: "Str", B: "String"}, Copy[string]())

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

	testTwoWayConversion(t, &bi, from, expectedTo)
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

	bi := StructBijection[RiverExample, YamlExample]{}

	// Test inverting a bijection too
	inverted := Inverted[int32, int64](Cast[int32, int64]())
	BindField(&bi, Names{A: "TestRiver", B: "TestYAML"}, inverted)
	BindField(&bi, Names{A: "UInt", B: "UIntValue"}, Cast[uint64, float64]())
	BindField(&bi, Names{A: "Str", B: "Bytes"}, Cast[string, []byte]())
	BindField(&bi, Names{A: "Bytes", B: "Str"}, Cast[[]byte, string]())

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

	testTwoWayConversion(t, &bi, from, expectedTo)
}

func testTwoWayConversion[A any, B any](t *testing.T, bi *StructBijection[A, B], from A, expectedTo B) {
	var to B
	err := bi.ConvertAToB(&from, &to)
	require.NoError(t, err)
	require.Equal(t, expectedTo, to)

	var reversed A
	err = bi.ConvertBToA(&to, &reversed)
	require.NoError(t, err)
	require.Equal(t, from, reversed)
}
