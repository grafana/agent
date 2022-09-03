package encoding

import (
	"testing"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	obj := &t1{
		Age: 1,
	}
	val := value.Encode(obj)
	tags := rivertags.Get(val.Reflect().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Age), tags[0])
	require.NoError(t, err)
	require.True(t, af.Name == age)
	require.True(t, af.valueField.Type == number)
	require.True(t, af.valueField.Value == 1)
}

func TestNested(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	type parent struct {
		Person *t1 `river:"person,attr"`
	}

	obj := &parent{Person: &t1{Age: 1}}
	val := value.Encode(obj)
	tags := rivertags.Get(val.Reflect().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Person), tags[0])
	require.NoError(t, err)
	require.True(t, af.Name == "person")
	require.True(t, af.structField.Type == object)
	require.Len(t, af.structField.Value, 1)
	require.True(t, af.structField.Value[0].Key == age)
	require.True(t, af.structField.Value[0].Value.(*ValueField).Value == 1)
	require.True(t, af.structField.Value[0].Value.(*ValueField).Type == number)
}
