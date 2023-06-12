package common

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// ConvertTargets implements [builder.Tokenizer]. This allows us to set
// component.Arguments with an implementation that can be tokenized with
// custom behaviour for converting.
type ConvertTargets struct {
	Targets []discovery.Target
}

var _ builder.Tokenizer = ConvertTargets{}
var _ river.Capsule = ConvertTargets{}

func (f ConvertTargets) RiverCapsule() {}
func (f ConvertTargets) RiverTokenize() []builder.Token {
	var toks []builder.Token

	targetCount := 0
	for _, targetMap := range f.Targets {
		targetCount += len(targetMap)
	}

	if targetCount == 0 {
		toks = append(toks, builder.Token{Tok: token.LBRACK})
		toks = append(toks, builder.Token{Tok: token.RBRACK})
		return toks
	}

	if targetCount > 1 {
		toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "concat"})
		toks = append(toks, builder.Token{Tok: token.LPAREN})
	}

	for ix, targetMap := range f.Targets {
		for key, target := range targetMap {
			if key == "__address__" {
				toks = append(toks, builder.Token{Tok: token.LBRACK})
				toks = append(toks, builder.Token{Tok: token.LCURLY})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: key})
				toks = append(toks, builder.Token{Tok: token.ASSIGN})
				toks = append(toks, builder.Token{Tok: token.STRING, Lit: `"` + target + `"`})
				toks = append(toks, builder.Token{Tok: token.COMMA})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
				toks = append(toks, builder.Token{Tok: token.RCURLY})
				toks = append(toks, builder.Token{Tok: token.RBRACK})
			} else {
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: key})
			}

			if ix != len(f.Targets)-1 {
				toks = append(toks, builder.Token{Tok: token.COMMA})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
			}
		}
	}

	if targetCount > 1 {
		toks = append(toks, builder.Token{Tok: token.RPAREN})
	}

	return toks
}
