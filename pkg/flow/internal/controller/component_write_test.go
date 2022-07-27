package controller

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Import test components
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/stretchr/testify/require"
)

func TestWriteComponent(t *testing.T) {
	config := `
		testcomponents.passthrough "example" {
			input = "Hello, world!"
		}
	`
	blocks := loadFile(t, []byte(config))

	cn := NewComponentNode(ComponentGlobals{
		Logger:          log.NewNopLogger(),
		DataPath:        t.TempDir(),
		OnExportsChange: func(cn *ComponentNode) { /* no-op */ },
	}, blocks[0])

	// Evaluate the component so we're sure it's built
	err := cn.Evaluate(nil)
	require.NoError(t, err)

	outBlock := WriteComponent(cn, false)
	actual := marshalBlock(outBlock)

	expect := `
testcomponents.passthrough "example" {
	input = "Hello, world!"

	// Exported fields:
	output = "Hello, world!"
}`

	// Remove leading and trailing whitespace so we don't have to get too picky
	// about how we format the expected string.
	expect = strings.TrimSpace(expect)
	actual = strings.TrimSpace(actual)
	require.Equal(t, expect, actual)
}

func TestWriteComponent_DebugInfo(t *testing.T) {
	config := `
		testcomponents.passthrough "example" {
			input = "Hello, world!"
		}
	`

	blocks := loadFile(t, []byte(config))

	cn := NewComponentNode(ComponentGlobals{
		Logger:          log.NewNopLogger(),
		DataPath:        t.TempDir(),
		OnExportsChange: func(cn *ComponentNode) { /* no-op */ },
	}, blocks[0])

	// Evaluate the component so we're sure it's built
	err := cn.Evaluate(nil)
	require.NoError(t, err)

	outBlock := WriteComponent(cn, true)
	actual := marshalBlock(outBlock)

	expect := fmt.Sprintf(`
testcomponents.passthrough "example" {
	input = "Hello, world!"

	// Exported fields:
	output = "Hello, world!"

	// Debug info:
	health {
		state       = "healthy"
		message     = "component evaluated"
		update_time = %q
	}

	status {
		component_version = "v0.1-beta.0"
	}
}`, cn.evalHealth.UpdateTime.Format(time.RFC3339Nano))

	// Remove leading and trailing whitespace so we don't have to get too picky
	// about how we format the expected string.
	expect = strings.TrimSpace(expect)
	actual = strings.TrimSpace(actual)
	require.Equal(t, expect, actual)
}

func loadFile(t *testing.T, bb []byte) []*ast.BlockStmt {
	file, err := parser.ParseFile(t.Name(), bb)
	require.NoError(t, err)

	var blocks []*ast.BlockStmt

	for _, stmt := range file.Body {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			blocks = append(blocks, stmt)
		default:
			require.FailNow(t, "%s: non-block statement unexpected", ast.StartPos(stmt).Position())
		}
	}

	return blocks
}

func marshalBlock(b *builder.Block) string {
	f := builder.NewFile()
	f.Body().AppendBlock(b)
	return string(f.Bytes())
}
