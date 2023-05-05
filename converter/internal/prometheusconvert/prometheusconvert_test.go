package prometheusconvert_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/stretchr/testify/require"
)

const (
	inputSuffix  = ".in.yaml"
	outputSuffix = ".out.river"
)

func Test(t *testing.T) {
	filepath.WalkDir("testdata", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, inputSuffix) {
			inputFile := path
			expectFile := strings.TrimSuffix(path, inputSuffix) + outputSuffix

			inputBytes, err := os.ReadFile(inputFile)
			require.NoError(t, err)
			expectBytes, err := os.ReadFile(expectFile)
			require.NoError(t, err)

			caseName := filepath.Base(path)
			caseName = strings.TrimSuffix(caseName, inputSuffix)

			t.Run(caseName, func(t *testing.T) {
				testConverter(t, inputBytes, expectBytes)
				testConverter2(t, inputBytes, expectBytes)
				// testConverter3(t, inputBytes, expectBytes)
			})
		}

		return nil
	})
}

func testConverter(t *testing.T, input, expect []byte) {
	t.Helper()

	actual, err := prometheusconvert.Convert(input)
	require.NoError(t, err)
	require.Equal(t, string(expect), string(actual)+"\n")
}

func testConverter2(t *testing.T, input, expect []byte) {
	t.Helper()

	actual, err := prometheusconvert.Convert2(input)
	require.NoError(t, err)
	require.Equal(t, string(expect), string(actual)+"\n")
}

func testConverter3(t *testing.T, input, expect []byte) {
	t.Helper()

	actual, err := prometheusconvert.Convert3(input)
	require.NoError(t, err)
	require.Equal(t, string(expect), string(actual)+"\n")
}
