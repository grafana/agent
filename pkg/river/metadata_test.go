package river

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	type testSimple struct {
		LogLevel string `river:"log_level,attr"`
	}

	f := &TagField{Fields: make([]*TagField, 0)}
	err := generateField(f, &testSimple{})
	require.NoError(t, err)
	require.NotNil(t, f)
	//require.True(t, f.Name == "test")
	require.True(t, f.Fields[0].Name == "log_level")
	require.True(t, f.Fields[0].DataType == "string")
	require.True(t, f.Fields[0].IsAttribute == true)
}

func TestArray(t *testing.T) {

	type test struct {
		LogLevels []string `river:"log_level,attr"`
	}
	f := &TagField{Fields: make([]*TagField, 0)}
	err := generateField(f, &test{})
	require.NoError(t, err)
	require.NotNil(t, f)
	//require.True(t, f.Name == "test")
	require.True(t, f.Fields[0].Name == "log_level")
	require.True(t, f.Fields[0].DataType == "array")
	require.True(t, f.Fields[0].IsArray == true)
	require.True(t, f.Fields[0].ArrayType == "string")
	require.True(t, f.Fields[0].IsAttribute == true)
}

func TestMap(t *testing.T) {
	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	f := &TagField{Fields: make([]*TagField, 0)}
	err := generateField(f, &test{})
	require.NoError(t, err)
	require.NotNil(t, f)
	//	require.True(t, f.Name == "test")
	require.True(t, f.Fields[0].Name == "log_level")
	require.True(t, f.Fields[0].DataType == "map")
	require.True(t, f.Fields[0].IsMap == true)
	require.True(t, f.Fields[0].MapKeyType == "string")
	require.True(t, f.Fields[0].MapValueType == "number")
	require.True(t, f.Fields[0].IsAttribute == true)
}

func TestChildren(t *testing.T) {

	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	type parent struct {
		T test `river:"child,block"`
	}
	f := &TagField{Fields: make([]*TagField, 0)}
	err := generateField(f, &parent{})
	require.NoError(t, err)
	require.NotNil(t, f)
	child := f.Fields[0]
	require.True(t, child.Name == "child")
	require.True(t, child.DataType == "object")
	require.True(t, child.IsMap == false)

	subchild := child.Fields[0]
	require.True(t, subchild.Name == "log_level")
	require.True(t, subchild.DataType == "map")
	require.True(t, subchild.IsMap == true)
	require.True(t, subchild.MapKeyType == "string")
	require.True(t, subchild.MapValueType == "number")
	require.True(t, subchild.IsAttribute == true)
}
