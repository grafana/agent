package prometheusconvert_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/stretchr/testify/require"
)

const (
	promSuffix  = ".yaml"
	flowSuffix  = ".river"
	diagsSuffix = ".diags"
)

func TestConvert(t *testing.T) {
	filepath.WalkDir("testdata", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, promSuffix) {
			inputFile := path
			inputBytes, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			caseName := filepath.Base(path)
			caseName = strings.TrimSuffix(caseName, promSuffix)

			t.Run(caseName, func(t *testing.T) {
				actual, diags := prometheusconvert.Convert(inputBytes)

				// Skip Info level diags for this testing
				diags.RemoveDiagsBySeverity(diag.SeverityLevelInfo)

				// Generate an HTML report
				err := diags.GenerateReport(diag.HTML, "/mnt/c/workspace/html/"+caseName+"-diagnostics.html")
				require.NoError(t, err)

				// Generate a text report
				err = diags.GenerateReport(diag.Text, "/mnt/c/workspace/html/"+caseName+"-diagnostics.txt")
				require.NoError(t, err)

				// Generate a text report
				// err = diags.GenerateReport(diag.Text, "/home/erik/go/src/agent/converter/internal/prometheusconvert/testdata/"+caseName+".diags")
				// require.NoError(t, err)

				expectedDiags := parseDiags(t, path)
				for ix, diag := range diags {
					if len(expectedDiags) > ix {
						require.Equal(t, expectedDiags[ix], diag.String())
					} else {
						require.Fail(t, "unexpected diag count reach for diag: "+diag.String())
					}
				}

				// If we expect more diags than we got
				if len(expectedDiags) > len(diags) {
					require.Fail(t, "missing expected diag: "+expectedDiags[len(diags)])
				}

				outputFile := strings.TrimSuffix(path, promSuffix) + flowSuffix
				if _, err := os.Stat(outputFile); err == nil {
					outputBytes, err := os.ReadFile(outputFile)
					require.NoError(t, err)
					require.Equal(t, string(normalizeLineEndings(outputBytes)), string(normalizeLineEndings(actual)))
				}
			})
		}

		return nil
	})
}

// Replace '\r\n' with '\n'
func normalizeLineEndings(data []byte) []byte {
	normalized := bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
	return normalized
}

func parseDiags(t *testing.T, path string) []string {
	expectedDiags := []string{}
	diagsFile := strings.TrimSuffix(path, promSuffix) + diagsSuffix
	if _, err := os.Stat(diagsFile); err == nil {
		errorBytes, err := os.ReadFile(diagsFile)
		require.NoError(t, err)
		errorsString := string(normalizeLineEndings(errorBytes))
		expectedDiags = strings.Split(errorsString, "\n")

		// Some error messages have \n in them and need this
		for ix := range expectedDiags {
			expectedDiags[ix] = strings.ReplaceAll(expectedDiags[ix], "\\n", "\n")
		}
	}

	return expectedDiags
}
