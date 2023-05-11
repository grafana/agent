package flow

import (
	"strings"
	"testing"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/stretchr/testify/require"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

func TestReadFile(t *testing.T) {
	content := `
		testcomponents.tick "ticker_a" {
			frequency = "1s"
		}

		testcomponents.passthrough "static" {
			input = "hello, world!"
		}
	`

	f, err := ParseSource(t.Name(), []byte(content))
	require.NoError(t, err)
	require.NotNil(t, f)

	require.Len(t, f.components, 2)
	require.Equal(t, "testcomponents.tick.ticker_a", getBlockID(f.components[0]))
	require.Equal(t, "testcomponents.passthrough.static", getBlockID(f.components[1]))
}

func TestReadFileWithConfigBlock(t *testing.T) {
	content := `
        logging {
		    log_format = "json"
		}

		testcomponents.tick "ticker_a" {
			frequency = "1s"
		}
	`

	f, err := ParseSource(t.Name(), []byte(content))
	require.NoError(t, err)
	require.NotNil(t, f)

	require.Len(t, f.components, 1)
	require.Equal(t, "testcomponents.tick.ticker_a", getBlockID(f.components[0]))
	require.Len(t, f.configBlocks, 1)
	require.Equal(t, "logging", getBlockID(f.configBlocks[0]))
}

func TestReadFile_Defaults(t *testing.T) {
	f, err := ParseSource(t.Name(), []byte(``))
	require.NotNil(t, f)
	require.NoError(t, err)

	require.Len(t, f.components, 0)
}

func getBlockID(b *ast.BlockStmt) string {
	var parts []string
	parts = append(parts, b.Name...)
	if b.Label != "" {
		parts = append(parts, b.Label)
	}
	return strings.Join(parts, ".")
}
