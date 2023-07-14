package test_common

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/agent/converter/diag"
	"github.com/stretchr/testify/require"
)

const (
	flowSuffix  = ".river"
	diagsSuffix = ".diags"
)

func TestDirectory(t *testing.T, folderPath string, sourceSuffix string, convert func(in []byte) ([]byte, diag.Diagnostics)) {
	require.NoError(t, filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, sourceSuffix) {
			inputFile := path
			inputBytes, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			caseName := filepath.Base(path)
			caseName = strings.TrimSuffix(caseName, sourceSuffix)

			t.Run(caseName, func(t *testing.T) {
				actual, diags := convert(inputBytes)

				// Skip Info level diags for this testing
				diags.RemoveDiagsBySeverity(diag.SeverityLevelInfo)

				expectedDiags := parseDiags(t, strings.TrimSuffix(path, sourceSuffix)+diagsSuffix)
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

				outputFile := strings.TrimSuffix(path, sourceSuffix) + flowSuffix
				if _, err := os.Stat(outputFile); err == nil {
					outputBytes, err := os.ReadFile(outputFile)
					require.NoError(t, err)
					fmt.Println("============== ACTUAL =============")
					fmt.Println(string(normalizeLineEndings(actual)))
					fmt.Println("===================================")
					require.Equal(t, string(normalizeLineEndings(outputBytes)), string(normalizeLineEndings(actual)))
				}
			})
		}

		return nil
	}))
}

// Replace '\r\n' with '\n'
func normalizeLineEndings(data []byte) []byte {
	normalized := bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
	return normalized
}

func parseDiags(t *testing.T, diagsFile string) []string {
	expectedDiags := []string{}
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
