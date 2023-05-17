package prometheusconvert_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
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

	actual, err := prometheusconvert.Convert(input)

	require.NoError(t, err)
	require.Equal(t, string(normalizeLineEndings(expect)), string(normalizeLineEndings(actual))+"\n")

}

// Replace '\r\n' with '\n'
func normalizeLineEndings(data []byte) []byte {
	normalized := bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
	return normalized
}

func TestFlowParsing(t *testing.T) {
	filepath.WalkDir("testdata", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, flowSuffix) {
			caseName := filepath.Base(path)
			caseName = strings.TrimSuffix(caseName, flowSuffix)

			t.Run(caseName, func(t *testing.T) {
				inputBytes, err := os.ReadFile(path)
				require.NoError(t, err)

				file, err := flow.ReadFile(path, inputBytes)
				require.NoError(t, err)

				f := flow.New(testOptions(t))
				err = f.LoadFile(file, nil)
				require.NoError(t, err)
			})
		}

		return nil
	})
}

func testOptions(t *testing.T) flow.Options {
	t.Helper()

	s, err := logging.WriterSink(os.Stderr, logging.DefaultSinkOptions)
	require.NoError(t, err)

	c := &cluster.Clusterer{Node: cluster.NewLocalNode("")}

	return flow.Options{
		LogSink:   s,
		DataPath:  t.TempDir(),
		Reg:       nil,
		Clusterer: c,
	}
}
