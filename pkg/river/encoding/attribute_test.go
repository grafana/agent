package encoding

import (
	"encoding/json"
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
	reqString := `{"name":"age","type":"attr","value":{"type":"number","value":1}}`
	val := value.Encode(obj)
	tags := rivertags.Get(val.Reflect().Type())
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
	tags := rivertags.Get(val.Reflect().Type())
	require.Len(t, tags, 1)

	af, err := newAttribute(value.Encode(obj.Person), tags[0])
	require.NoError(t, err)
	bb, err := json.Marshal(af)
	require.NoError(t, err)
	require.JSONEq(t, reqString, string(bb))
}
