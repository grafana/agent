package prometheusconvert

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	promSuffix = ".yaml"
	flowSuffix = ".river"
)

func TestConvert(t *testing.T) {
	filepath.WalkDir("testdata", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, promSuffix) {
			inputFile := path
			expectFile := strings.TrimSuffix(path, promSuffix) + flowSuffix

			inputBytes, err := os.ReadFile(inputFile)
			require.NoError(t, err)
			expectBytes, err := os.ReadFile(expectFile)
			require.NoError(t, err)

			caseName := filepath.Base(path)
			caseName = strings.TrimSuffix(caseName, promSuffix)

			t.Run(caseName, func(t *testing.T) {
				testConverter(t, inputBytes, expectBytes)
			})
		}

		return nil
	})
}

func testConverter(t *testing.T, input, expect []byte) {
	t.Helper()

	actual, err := Convert(input)

	require.NoError(t, err)
	require.Equal(t, string(normalizeLineEndings(expect)), string(normalizeLineEndings(actual))+"\n")
}

// Replace '\r\n' with '\n'
func normalizeLineEndings(data []byte) []byte {
	normalized := bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
	return normalized
}
