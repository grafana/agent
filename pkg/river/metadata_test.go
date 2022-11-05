package river

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	type testSimple struct {
		LogLevel string `river:"log_level,attr"`
	}

	md := MetadataDict{Types: make([]DataType, 0)}
	_, err := md.generateField("testSimple", reflect.TypeOf(testSimple{}))
	require.NoError(t, err)
	require.Len(t, md.Types, 1)
	require.True(t, md.Types[0].Name == "testSimple")
	require.True(t, md.Types[0].Fields[0].Name == "log_level")
	require.True(t, md.Types[0].Fields[0].IsAttribute)
	require.False(t, md.Types[0].Fields[0].IsBlock)
	require.False(t, md.Types[0].Fields[0].IsMap)
	require.True(t, md.Types[0].Fields[0].DataType == "string")

}

func TestArray(t *testing.T) {

	type test struct {
		LogLevels []string `river:"log_level,attr"`
	}
	md := MetadataDict{Types: make([]DataType, 0)}
	_, err := md.generateField("test", reflect.TypeOf(test{}))
	require.NoError(t, err)
	require.True(t, md.Types[0].Name == "test")
	require.True(t, md.Types[0].Fields[0].Name == "log_level")
	require.True(t, md.Types[0].Fields[0].IsAttribute)
	require.False(t, md.Types[0].Fields[0].IsBlock)
	require.True(t, md.Types[0].Fields[0].IsArray)
	require.True(t, md.Types[0].Fields[0].DataType == "array")
	require.True(t, md.Types[0].Fields[0].ArrayType == "string")
}

func TestMap(t *testing.T) {
	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	md := MetadataDict{Types: make([]DataType, 0)}
	_, err := md.generateField("test", reflect.TypeOf(test{}))
	require.NoError(t, err)
	require.True(t, md.Types[0].Name == "test")
	require.True(t, md.Types[0].Fields[0].Name == "log_level")
	require.True(t, md.Types[0].Fields[0].IsAttribute)
	require.False(t, md.Types[0].Fields[0].IsBlock)
	require.False(t, md.Types[0].Fields[0].IsArray)
	require.True(t, md.Types[0].Fields[0].IsMap)
	require.True(t, md.Types[0].Fields[0].DataType == "map")
	require.True(t, md.Types[0].Fields[0].MapKeyType == "string")
	require.True(t, md.Types[0].Fields[0].MapValueType == "number")
}

func TestChildren(t *testing.T) {

	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	type parent struct {
		T test `river:"child,block"`
	}
	md := MetadataDict{Types: make([]DataType, 0)}
	_, err := md.generateField("parent", reflect.TypeOf(parent{}))
	require.NoError(t, err)
	require.NoError(t, err)
	require.Len(t, md.Types, 2)
	found, p := find(md, "parent")
	require.True(t, found)
	found, c := find(md, "child")
	require.True(t, found)
	require.True(t, p.Name == "parent")
	require.Len(t, p.Fields, 1)
	require.True(t, p.Fields[0].Name == "child")
	require.True(t, p.Fields[0].IsBlock)
	require.True(t, p.Fields[0].DataType == "child")

	require.True(t, c.Name == "child")
	require.True(t, c.Fields[0].IsMap)
	require.True(t, c.Fields[0].DataType == "map")
	require.True(t, c.Fields[0].MapKeyType == "string")
	require.True(t, c.Fields[0].MapValueType == "number")
}

func TestUnique(t *testing.T) {

	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	type parent struct {
		T  test `river:"child,block"`
		T2 test `river:"child2,block"`
	}
	md := MetadataDict{Types: make([]DataType, 0)}
	_, err := md.generateField("parent", reflect.TypeOf(&parent{}))
	require.NoError(t, err)
	require.NoError(t, err)
	require.Len(t, md.Types, 2)
	found, p := find(md, "parent")
	require.True(t, found)
	found, c := find(md, "child")
	require.True(t, found)
	require.True(t, p.Name == "parent")
	require.Len(t, p.Fields, 2)
	require.True(t, p.Fields[0].Name == "child")
	require.True(t, p.Fields[0].IsBlock)
	require.True(t, p.Fields[0].DataType == "child")

	require.True(t, p.Fields[1].Name == "child2")
	require.True(t, p.Fields[1].IsBlock)
	require.True(t, p.Fields[1].DataType == "child")

	require.True(t, c.Name == "child")
	require.True(t, c.Fields[0].IsMap)
	require.True(t, c.Fields[0].DataType == "map")
	require.True(t, c.Fields[0].MapKeyType == "string")
	require.True(t, c.Fields[0].MapValueType == "number")

	found, _ = find(md, "child2")
	require.False(t, found)
}

func find(md MetadataDict, name string) (bool, DataType) {
	for _, x := range md.Types {
		if x.Name == name {
			return true, x
		}
	}
	return false, DataType{}
}
