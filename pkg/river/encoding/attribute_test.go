package encoding

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/component/discovery"
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
	reqString := `{"name":"age","type":"attr","value":{"type":"number","value":1}}`
	val := value.Encode(obj)
	tags := rivertags.Get(val.GetValue().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Age), tags[0])
	require.NoError(t, err)
	bb, err := json.Marshal(af)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}

func TestSimpleZeroVal(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	obj := &t1{
		Age: 0,
	}
	reqString := `{"name":"age","type":"attr","value":{"type":"number","value":0}}`
	val := value.Encode(obj)
	tags := rivertags.Get(val.GetValue().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Age), tags[0])
	require.NoError(t, err)
	bb, err := json.Marshal(af)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}

func TestNested(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	type parent struct {
		Person *t1 `river:"person,attr"`
	}
	reqString := `{"name":"person","type":"attr","value":{"type":"object","value":[{"value":{"type":"number","value":1},"key":"age"}]}}`

	obj := &parent{Person: &t1{Age: 1}}
	val := value.Encode(obj)
	tags := rivertags.Get(val.GetValue().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Person), tags[0])
	require.NoError(t, err)
	bb, err := json.Marshal(af)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}

func TestNestedZeroVal(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	type parent struct {
		Person *t1 `river:"person,attr"`
	}
	reqString := `{"name":"person","type":"attr","value":{"type":"object","value":[{"value":{"type":"number","value":0},"key":"age"}]}}`

	obj := &parent{Person: &t1{Age: 0}}
	val := value.Encode(obj)
	tags := rivertags.Get(val.GetValue().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Person), tags[0])
	require.NoError(t, err)
	bb, err := json.Marshal(af)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}

func TestDiscovery(t *testing.T) {
	type t1 struct {
		Targets []discovery.Target `river:"targets,attr"`
	}
	testObj := &t1{
		Targets: make([]discovery.Target, 0),
	}
	testObj.Targets = append(testObj.Targets, discovery.Target{"t": "test"})
	val := value.Encode(testObj)
	tags := rivertags.Get(val.GetValue().Type())
	attr, err := newAttribute(value.Encode(testObj.Targets), tags[0])
	require.NoError(t, err)
	require.True(t, attr.Name == "targets")
	require.True(t, attr.hasValue())
}

func TestDiscoveryNil(t *testing.T) {
	type t1 struct {
		Targets []discovery.Target `river:"targets,attr"`
	}
	testObj := &t1{
		Targets: make([]discovery.Target, 0),
	}
	val := value.Encode(testObj)
	tags := rivertags.Get(val.GetValue().Type())
	attr, err := newAttribute(value.Encode(testObj.Targets), tags[0])
	require.NoError(t, err)
	require.True(t, attr.Name == "targets")
	require.False(t, attr.hasValue())
}
