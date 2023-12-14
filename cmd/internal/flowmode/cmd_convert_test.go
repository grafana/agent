package flowmode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseExtraArgs(t *testing.T) {
	type testCase struct {
		name      string
		extraArgs string
		expected  []string
	}

	var testCases = []testCase{
		{
			name:      "full",
			extraArgs: "--key=value",
			expected:  []string{"--key", "value"},
		},
		{
			name:      "shorthand",
			extraArgs: "-k=value",
			expected:  []string{"-k", "value"},
		},
		{
			name:      "bool",
			extraArgs: "--boolVariable",
			expected:  []string{"--boolVariable"},
		},
		{
			name:      "spaced",
			extraArgs: "--key value",
			expected:  []string{"--key", "value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := parseExtraArgs(tc.extraArgs)
			require.NoError(t, err)
			require.Equal(t, tc.expected, res)
		})
	}
}
