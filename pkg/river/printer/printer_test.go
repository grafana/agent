package printer_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
	filepath.WalkDir("testdata", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".in") {
			inputFile := path
			expectFile := strings.TrimSuffix(path, ".in") + ".expect"

			inputBB, err := os.ReadFile(inputFile)
			require.NoError(t, err)
			expectBB, err := os.ReadFile(expectFile)
			require.NoError(t, err)

			caseName := filepath.Base(path)
			caseName = strings.TrimSuffix(caseName, ".in")

			t.Run(caseName, func(t *testing.T) {
				testPrinter(t, inputBB, expectBB)
			})
		}

		return nil
	})
}

func testPrinter(t *testing.T, input, expect []byte) {
	f, err := parser.ParseFile(t.Name()+".rvr", input)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, printer.Fprint(&buf, f))

	trimmed := strings.TrimRightFunc(string(expect), unicode.IsSpace)
	require.Equal(t, trimmed, buf.String(), "%s", buf.String())
}
