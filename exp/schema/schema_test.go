package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	type sts struct {
		Name string `river:"name,attr,optional"`
		Age  int    `river:"age,attr,optional"`
	}
	s := sts{
		Name: "bob",
		Age:  16,
	}
	sch := NewSchema("t", "1")
	err := sch.AddComponent("comp1", "", s, nil)
	require.NoError(t, err)
	require.Len(t, sch.Json.Components, 1)
	require.Len(t, sch.Json.Components[0].Arguments, 2)

	require.True(t, containsName(sch.Json.Components[0].Arguments, "name"))
	require.True(t, containsName(sch.Json.Components[0].Arguments, "age"))
}

func TestArray(t *testing.T) {
	type address struct {
		City string `river:"city,attr,optional"`
	}
	type person struct {
		Name    string    `river:"name,attr,optional"`
		Age     int       `river:"age,attr,optional"`
		Address []address `river:"address,block,optional"`
	}
	s := person{
		Name: "bob",
		Age:  16,
		Address: []address{
			{
				City: "new york",
			},
		},
	}
	sch := NewSchema("t", "1")
	err := sch.AddComponent("comp1", "", s, nil)
	require.NoError(t, err)
	require.Len(t, sch.Json.Components, 1)
	require.Len(t, sch.Json.Components[0].Arguments, 3)

	require.True(t, containsName(sch.Json.Components[0].Arguments, "name"))
	require.True(t, containsName(sch.Json.Components[0].Arguments, "age"))

	addr, found := findType(sch.Json.Components[0].Arguments, "address")
	require.True(t, found)
	require.True(t, addr.Type == "array")
	require.True(t, addr.Name == "address")
	require.Len(t, addr.Children, 1)
	require.True(t, addr.Children[0].Name == "city")
}

func TestMap(t *testing.T) {
	type address struct {
		City string `river:"city,attr,optional"`
	}
	type person struct {
		Name    string             `river:"name,attr,optional"`
		Age     int                `river:"age,attr,optional"`
		Address map[string]address `river:"address,block,optional"`
	}
	s := person{
		Name: "bob",
		Age:  16,
		Address: map[string]address{
			"one": {
				City: "new york",
			},
		},
	}
	sch := NewSchema("t", "1")
	err := sch.AddComponent("comp1", "", s, nil)
	require.NoError(t, err)
	require.Len(t, sch.Json.Components, 1)
	require.Len(t, sch.Json.Components[0].Arguments, 3)

	require.True(t, containsName(sch.Json.Components[0].Arguments, "name"))
	require.True(t, containsName(sch.Json.Components[0].Arguments, "age"))

	addr, found := findType(sch.Json.Components[0].Arguments, "address")
	require.True(t, found)
	require.True(t, addr.Type == "map")
	require.True(t, addr.Name == "address")
	require.Len(t, addr.Children, 1)
	require.True(t, addr.Children[0].Name == "city")
}

func TestStruct(t *testing.T) {
	type address struct {
		City string `river:"city,attr,optional"`
	}
	type person struct {
		Name    string  `river:"name,attr,optional"`
		Age     int     `river:"age,attr,optional"`
		Address address `river:"address,block,optional"`
	}
	s := person{
		Name: "bob",
		Age:  16,
		Address: address{
			City: "new york",
		},
	}
	sch := NewSchema("t", "1")
	err := sch.AddComponent("comp1", "", s, nil)
	require.NoError(t, err)
	require.Len(t, sch.Json.Components, 1)
	require.Len(t, sch.Json.Components[0].Arguments, 3)

	require.True(t, containsName(sch.Json.Components[0].Arguments, "name"))
	require.True(t, containsName(sch.Json.Components[0].Arguments, "age"))

	addr, found := findType(sch.Json.Components[0].Arguments, "address")
	require.True(t, found)
	require.True(t, addr.Type == "object")
	require.True(t, addr.Name == "address")
	require.Len(t, addr.Children, 1)
	require.True(t, addr.Children[0].Name == "city")
}

func findType(children []Type, name string) (Type, bool) {
	for _, k := range children {
		if k.Name == name {
			return k, true
		}
	}
	return Type{}, false
}

func containsName(children []Type, name string) bool {
	for _, k := range children {
		if k.Name == name {
			return true
		}
	}
	return false
}
