package metadata

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_inferMetadata(t *testing.T) {
	tests := []struct {
		name     string
		expected Metadata
	}{
		{
			name:     "discovery.dns",
			expected: Metadata{Outputs: []DataType{DataTypeTargets}},
		},
		{
			name: "discovery.relabel",
			expected: Metadata{
				Accepts: []DataType{DataTypeTargets},
				Outputs: []DataType{DataTypeTargets},
			},
		},
		{
			name:     "loki.echo",
			expected: Metadata{Accepts: []DataType{DataTypeLokiLogs}},
		},
		{
			name: "loki.source.file",
			expected: Metadata{
				Accepts: []DataType{DataTypeTargets},
				Outputs: []DataType{DataTypeLokiLogs},
			},
		},
		{
			name: "loki.process",
			expected: Metadata{
				Accepts: []DataType{DataTypeLokiLogs},
				Outputs: []DataType{DataTypeLokiLogs},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ForComponent(tt.name)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
