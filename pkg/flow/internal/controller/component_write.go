package controller

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
)

// WriteComponent generates an hclwrite Block from a component. Health and
// debug info will be included if debugInfo is true.
func WriteComponent(cn *ComponentNode, debugInfo bool) *hclwrite.Block {
	var (
		id = cn.ID()

		blockName = id[0]
		labels    = id[1:]
	)

	b := hclwrite.NewBlock(blockName, labels)

	if args := cn.Arguments(); args != nil {
		gohcl.EncodeIntoBody(args, b.Body())
	}

	if exports := cn.Exports(); exports != nil {
		b.Body().AppendUnstructuredTokens(hclwrite.Tokens{
			{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
			{Type: hclsyntax.TokenComment, Bytes: []byte("// Exported fields:\n")},
		})
		gohcl.EncodeIntoBody(exports, b.Body())
	}

	if debugInfo {
		b.Body().AppendUnstructuredTokens(hclwrite.Tokens{
			{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
			{Type: hclsyntax.TokenComment, Bytes: []byte("// Debug info:\n")},
		})

		b.Body().AppendBlock(gohcl.EncodeAsBlock(cn.CurrentHealth(), "health"))

		if di := cn.DebugInfo(); di != nil {
			b.Body().AppendBlock(gohcl.EncodeAsBlock(di, "status"))
		}
	}

	return b
}
