package prometheusconvert_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/river/diag"
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

				expectedDiags := diag.Diagnostics(nil)
				errorFile := strings.TrimSuffix(path, promSuffix) + errorsSuffix
				if _, err := os.Stat(errorFile); err == nil {
					errorBytes, err := os.ReadFile(errorFile)
					require.NoError(t, err)
					expectedDiags = parseErrors(t, errorBytes)
				}

				require.Equal(t, expectedDiags, diags)

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

func parseErrors(t *testing.T, errors []byte) diag.Diagnostics {
	var diags diag.Diagnostics

	errorsString := string(normalizeLineEndings(errors))
	splitErrors := strings.Split(errorsString, "\n")
	for _, error := range splitErrors {
		parsedError := strings.Split(error, " | ")
		if len(parsedError) != 2 {
			require.FailNow(t, "invalid error format")
		}

		severity, err := strconv.ParseInt(parsedError[0], 10, 8)
		require.NoError(t, err)

		// Some error messages have \n in them and need this
		errorMessage := strings.ReplaceAll(parsedError[1], "\\n", "\n")

		diags.Add(diag.Diagnostic{
			Severity: diag.Severity(severity),
			Message:  errorMessage,
		})
	}

	return diags
}
