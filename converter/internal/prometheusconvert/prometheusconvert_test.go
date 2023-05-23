package prometheusconvert_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	tt := []struct {
		name          string
		inputFile     string
		outputFile    string
		expectedDiags diag.Diagnostics
	}{
		{
			name:          "prometheus_agent",
			inputFile:     "prometheus_agent.yaml",
			outputFile:    "prometheus_agent.river",
			expectedDiags: nil,
		},
		{
			name:       "prometheus_agent_unsupported",
			inputFile:  "prometheus_agent_unsupported.yaml",
			outputFile: "prometheus_agent.river",
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.SeverityLevelWarn,
					Message:  "unsupported nomad_sd_config was provided",
				}},
		},
		{
			name:      "prometheus_agent_bad_config",
			inputFile: "prometheus_agent_bad_config.yaml",
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  "yaml: unmarshal errors:\n  line 7: field not_a_thing not found in type config.plain",
				}},
		},
		{
			name:      "prometheus_agent_broken_yaml",
			inputFile: "prometheus_agent_broken_yaml.yaml",
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  "yaml: line 18: did not find expected key",
				}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			inputBytes, err := os.ReadFile("testdata/" + tc.inputFile)
			require.NoError(t, err)

			actual, diags := prometheusconvert.Convert(inputBytes)
			require.Equal(t, tc.expectedDiags, diags)

			// If we expect errors, don't try to validate the output for this test
			if !tc.expectedDiags.HasErrors() {
				outputBytes, err := os.ReadFile("testdata/" + tc.outputFile)
				require.NoError(t, err)
				require.Equal(t, string(normalizeLineEndings(outputBytes)), string(normalizeLineEndings(actual))+"\n")
			}
		})
	}
}

// Replace '\r\n' with '\n'
func normalizeLineEndings(data []byte) []byte {
	normalized := bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
	return normalized
}
