package prometheusconvert_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/stretchr/testify/require"
)

const (
	promSuffix   = ".yaml"
	flowSuffix   = ".river"
	errorsSuffix = ".errors"
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

				expectedErrors := parseErrors(t, path)
				for ix, diag := range diags {
					if len(expectedErrors) > ix {
						require.Equal(t, expectedErrors[ix], diag.String())
					} else {
						require.Fail(t, "unexpected error count reach for error: "+diag.String())
					}
				}

				// If we expect more errors than we got
				if len(expectedErrors) > len(diags) {
					require.Fail(t, "missing expected error: "+expectedErrors[len(diags)])
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

func parseErrors(t *testing.T, path string) []string {
	expectedErrors := []string{}
	errorFile := strings.TrimSuffix(path, promSuffix) + errorsSuffix
	if _, err := os.Stat(errorFile); err == nil {
		errorBytes, err := os.ReadFile(errorFile)
		require.NoError(t, err)
		errorsString := string(normalizeLineEndings(errorBytes))
		expectedErrors = strings.Split(errorsString, "\n")

		// Some error messages have \n in them and need this
		for ix := range expectedErrors {
			expectedErrors[ix] = strings.ReplaceAll(expectedErrors[ix], "\\n", "\n")
		}
	}

	return expectedErrors
}
