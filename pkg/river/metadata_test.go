package river

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedefine(t *testing.T) {
	type Target map[string]string
	type testSimple struct {
		T []Target `river:"t,attr"`
	}
	md := NewMetaDict()
	_, err := md.generateField("testSimple", reflect.TypeOf(testSimple{}))
	require.NoError(t, err)
	valid, err := md.Verify()
	require.NoError(t, err)
	require.True(t, valid)
	require.Len(t, md.Types, 1)
	require.Len(t, md.NonRiverTypes, 6)

}

func TestMetadata(t *testing.T) {
	type testSimple struct {
		LogLevel string `river:"log_level,attr"`
	}

	md := NewMetaDict()
	_, err := md.generateField("testSimple", reflect.TypeOf(testSimple{}))
	require.NoError(t, err)
	require.Len(t, md.Types, 1)
	require.True(t, md.Types[0].Name == "testSimple")
	require.True(t, md.Types[0].Fields[0].Name == "log_level")
	require.True(t, md.Types[0].Fields[0].IsAttribute)
	require.False(t, md.Types[0].Fields[0].IsBlock)
	require.False(t, md.Types[0].Fields[0].IsMap)
	require.True(t, md.Types[0].Fields[0].DataType == "string")
	valid, _ := md.Verify()
	require.True(t, valid)
}

func TestArray(t *testing.T) {

	type test struct {
		LogLevels []string `river:"log_level,attr"`
	}
	md := NewMetaDict()
	_, err := md.generateField("test", reflect.TypeOf(test{}))
	require.NoError(t, err)
	require.True(t, md.Types[0].Name == "test")
	require.True(t, md.Types[0].Fields[0].Name == "log_level")
	require.True(t, md.Types[0].Fields[0].IsAttribute)
	require.False(t, md.Types[0].Fields[0].IsBlock)
	require.True(t, md.Types[0].Fields[0].IsArray)
	require.True(t, md.Types[0].Fields[0].DataType == "array")
	require.True(t, md.Types[0].Fields[0].ArrayType == "string")
	valid, _ := md.Verify()
	require.True(t, valid)
}

func TestMap(t *testing.T) {
	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	md := NewMetaDict()
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
	valid, _ := md.Verify()
	require.True(t, valid)
}

func TestChildren(t *testing.T) {

	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	type parent struct {
		T test `river:"child,block"`
	}
	md := NewMetaDict()
	_, err := md.generateField("parent", reflect.TypeOf(parent{}))
	require.NoError(t, err)
	require.NoError(t, err)
	require.Len(t, md.Types, 2)
	found, p := md.find("parent")
	require.True(t, found)
	found, c := md.find("child")
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
	valid, _ := md.Verify()
	require.True(t, valid)
}

func TestUnique(t *testing.T) {

	type test struct {
		LogLevels map[string]int `river:"log_level,attr"`
	}
	type parent struct {
		T  test `river:"child,block"`
		T2 test `river:"child2,block"`
	}
	md := NewMetaDict()
	_, err := md.generateField("parent", reflect.TypeOf(&parent{}))
	require.NoError(t, err)
	require.NoError(t, err)
	require.Len(t, md.Types, 2)
	found, p := md.find("parent")
	require.True(t, found)
	found, c := md.find("child")
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

	found, _ = md.find("child2")
	require.False(t, found)
	valid, _ := md.Verify()
	require.True(t, valid)
}
