package encoding

import (
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	tt := []struct {
		name    string
		testMap map[string]interface{}
	}{
		{
			name:    "Test Map Value",
			testMap: map[string]any{"testValue": "value"},
		},
		{
			name:    "Test Map Blank",
			testMap: map[string]any{"testBlank": ""},
		},
		{
			name:    "Test Map Null",
			testMap: map[string]any{"testNull": value.Null},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := newRiverMap(value.Encode(tc.testMap))
			require.NoError(t, err)
			require.True(t, mf.hasValue())
		})
	}
}
