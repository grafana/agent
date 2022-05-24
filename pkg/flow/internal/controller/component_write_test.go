package controller

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Import test components
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
)

func TestWriteComponent(t *testing.T) {
	config := `
		testcomponents "passthrough" "example" {
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
testcomponents "passthrough" "example" {
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
		testcomponents "passthrough" "example" {
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
testcomponents "passthrough" "example" {
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

func loadFile(t *testing.T, bb []byte) hcl.Blocks {
	file, diags := hclsyntax.ParseConfig(bb, t.Name(), hcl.InitialPos)
	if diags.HasErrors() {
		require.FailNow(t, diags.Error())
	}

	blockSchema := component.RegistrySchema()
	content, contentDiags := file.Body.Content(blockSchema)
	if contentDiags.HasErrors() {
		require.FailNow(t, contentDiags.Error())
	}

	return content.Blocks
}

func marshalBlock(b *hclwrite.Block) string {
	f := hclwrite.NewFile()
	f.Body().AppendBlock(b)
	return string(f.Bytes())
}
