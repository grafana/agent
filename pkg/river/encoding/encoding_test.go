package encoding_test

import (
	"io"
	"testing"

	"github.com/grafana/agent/pkg/river/encoding"
	"github.com/stretchr/testify/require"
)

func TestConvertRiverBodyToJSON_CapsuleValue(t *testing.T) {
	type Content struct {
		Field io.Closer `river:"writer,attr"`
	}

	out, err := encoding.ConvertRiverBodyToJSON(Content{Field: io.NopCloser(nil)})
	require.NoError(t, err)

	expect := `[
		{
			"name": "writer",
			"type": "attr",
			"value": {
				"type": "capsule",
				"value": "capsule(\"io.Closer\")"
			}
		}
	]`

	require.JSONEq(t, expect, string(out))
}

func TestConvertRiverBodyToJSON_BlockWithZeroValue(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	type parent struct {
		Person *t1 `river:"person,block"`
	}

	out, err := encoding.ConvertRiverBodyToJSON(parent{Person: &t1{Age: 0}})
	require.NoError(t, err)

	expect := `[{
		"name": "person",
		"type": "block",
		"body": [{
			"name": "age",
			"type": "attr",
			"value": {
				"type": "number",
				"value": 0
			}
		}]
	}]`

	require.JSONEq(t, expect, string(out))
}
