package auto

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type RiverExample struct {
	NestArrPtr []*NestedRiver `river:"nest_arr_ptr,block"`
	TestRiver  int            `river:"test,attr"`
	UInt       uint           `river:"uint_test,attr"`
	Str        string         `river:"str,attr"`
	Default    int
	Nested     NestedRiver   `river:"nested,block"`
	NestedPtr  *NestedRiver  `river:"nested_ptr,block"`
	StrArray   []string      `river:"str_arr,attr"`
	NestArr    []NestedRiver `river:"nest_arr,block"`
}

type NestedRiver struct {
	Int int    `river:"int,attr"`
	Str string `river:"str,attr"`
}

type YamlExample struct {
	TestYAML      int    `yaml:"test"`
	String        string `yaml:"str"`
	UIntValue     uint   `yaml:"uint_test,attr"`
	Default       int
	NestedYAML    NestedYAML    `yaml:"nested"`
	NestedYAMLPtr *NestedYAML   `yaml:"nested_ptr"`
	StrArr        []string      `yaml:"str_arr"`
	NestArray     []NestedYAML  `yaml:"nest_arr"`
	NestArrayPtr  []*NestedYAML `yaml:"nest_arr_ptr"`
}

type NestedYAML struct {
	Int int    `yaml:"int"`
	Str string `yaml:"str"`
}

func TestFoo(t *testing.T) {
	from := &RiverExample{
		TestRiver: 42,
		Str:       "foo",
		UInt:      31337,
		Default:   123,
		Nested: NestedRiver{
			Int: 321,
			Str: "hello",
		},
		NestedPtr: &NestedRiver{
			Int: 4321,
			Str: "hello ptr",
		},
		StrArray: []string{"foo", "bar"},
		NestArr: []NestedRiver{
			{Int: 1, Str: "1"},
			{Int: 2, Str: "2"},
		},
		NestArrPtr: []*NestedRiver{
			{Int: 3, Str: "3"},
			{Int: 4, Str: "4"},
		},
	}
	to := &YamlExample{}
	err := ConvertByFieldNames(from, to, RiverToYaml)
	require.NoError(t, err)
	require.Equal(t, 42, to.TestYAML)
	require.Equal(t, "foo", to.String)
	require.Equal(t, uint(31337), to.UIntValue)
	require.Equal(t, 123, to.Default)
	require.Equal(t, 321, to.NestedYAML.Int)
	require.Equal(t, "hello", to.NestedYAML.Str)
	require.Equal(t, 4321, to.NestedYAMLPtr.Int)
	require.Equal(t, "hello ptr", to.NestedYAMLPtr.Str)
	require.Equal(t, []string{"foo", "bar"}, to.StrArr)
	require.Equal(t, []NestedYAML{
		{Int: 1, Str: "1"},
		{Int: 2, Str: "2"},
	}, to.NestArray)
	require.Equal(t, []*NestedYAML{
		{Int: 3, Str: "3"},
		{Int: 4, Str: "4"},
	}, to.NestArrayPtr)
}
