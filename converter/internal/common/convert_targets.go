package common

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// ConvertTargets implements [builder.Tokenizer]. This allows us to set
// component.Arguments with an implementation that can be tokenized with
// custom behaviour for converting.
type ConvertTargets struct {
	Targets []discovery.Target
}

func (f ConvertTargets) RiverCapsule() {}
func (f ConvertTargets) RiverTokenize() []builder.Token {
	var toks []builder.Token
	if len(f.Targets) == 0 {
		toks = append(toks, builder.Token{Tok: token.LBRACK, Lit: ""})
		toks = append(toks, builder.Token{Tok: token.RBRACK, Lit: ""})
		return toks
	}

	// We are relying on each targetMap having exactly 1 target which we control
	// from the upstream converter code. We panic below if not.
	if len(f.Targets) > 1 {
		toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "concat"})
		toks = append(toks, builder.Token{Tok: token.LPAREN, Lit: ""})
	}

	for ix, targetMap := range f.Targets {
		if len(targetMap) != 1 {
			panic("unexpected number of targets received on a target map")
		}

		for key, target := range targetMap {
			if key == "__address__" {
				toks = append(toks, builder.Token{Tok: token.LBRACK, Lit: ""})
				toks = append(toks, builder.Token{Tok: token.LCURLY, Lit: ""})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: key})
				toks = append(toks, builder.Token{Tok: token.ASSIGN, Lit: ""})
				toks = append(toks, builder.Token{Tok: token.STRING, Lit: `"` + target + `"`})
				toks = append(toks, builder.Token{Tok: token.COMMA, Lit: ""})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
				toks = append(toks, builder.Token{Tok: token.RCURLY, Lit: ""})
				toks = append(toks, builder.Token{Tok: token.RBRACK, Lit: ""})
			} else {
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: key})
			}

			if ix != len(f.Targets)-1 {
				toks = append(toks, builder.Token{Tok: token.COMMA, Lit: ""})
				toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
			}
		}
	}

	if len(f.Targets) > 1 {
		toks = append(toks, builder.Token{Tok: token.RPAREN, Lit: ""})
	}

	return toks
}

var _ builder.Tokenizer = ConvertTargets{}
