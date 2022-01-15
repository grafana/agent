package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestSubsystemOptions_Unmarshal(t *testing.T) {

	tt := []struct {
		name        string
		in          string
		expectError string
	}{
		{
			name: "invalid integration",
			in: `
        invalidintegration: 
          autoscrape:
            enabled: true
      `,
			expectError: "line 2: field invalidintegration not found in type v2.plain",
		},
		{
			name: "invalid field",
			in: `
        test:
          invalidfield: true
      `,
			expectError: "line 2: field test not found in type v2.plain",
		},
		{
			name: "invalid v1 field",
			in: `
        windows_exporter:
          invalidfield: true
      `,
			expectError: "line 3: field invalidfield not found in type v2.plain",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			var so SubsystemOptions
			err := yaml.UnmarshalStrict([]byte(tc.in), &so)

			var te *yaml.TypeError
			require.ErrorAs(t, err, &te)
			require.Len(t, te.Errors, 1)
			require.Equal(t, tc.expectError, te.Errors[0])
		})
	}
}
