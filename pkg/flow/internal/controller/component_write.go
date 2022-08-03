package controller

import (
	"reflect"

	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// WriteComponent generates a token/builder Block from a component. Health and
// debug info will be included if debugInfo is true.
func WriteComponent(cn *ComponentNode, debugInfo bool) *builder.Block {
	b := builder.NewBlock(cn.getBlock().Name, cn.getBlock().Label)

	if args := cn.Arguments(); args != nil {
		b.Body().AppendFrom(args)
	}

	// We ignore zero value exports since the zero values for fields don't get
	// written back out to the user.
	if exports := cn.Exports(); exports != nil && !exportsZeroValue(exports) {
		b.Body().AppendTokens([]builder.Token{
			{Tok: token.LITERAL, Lit: "\n"},
			{Tok: token.COMMENT, Lit: "// Exported fields:"},
		})

		b.Body().AppendFrom(exports)
	}

	if debugInfo {
		b.Body().AppendTokens([]builder.Token{
			{Tok: token.LITERAL, Lit: "\n"},
			{Tok: token.COMMENT, Lit: "// Debug info:"},
		})

		healthBlock := builder.NewBlock([]string{"health"}, "")
		healthBlock.Body().AppendFrom(cn.CurrentHealth())
		b.Body().AppendBlock(healthBlock)

		if di := cn.DebugInfo(); di != nil {
			statusBlock := builder.NewBlock([]string{"status"}, "")
			statusBlock.Body().AppendFrom(di)
			b.Body().AppendBlock(statusBlock)
		}
	}

	return b
}

func exportsZeroValue(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}
