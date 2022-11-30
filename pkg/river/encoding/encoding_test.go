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
