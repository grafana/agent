package common

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// ConvertAppendable implements both the [builder.Tokenizer] and
// [storage.Appendable] interfaces. This allows us to set component.Arguments
// that leverage [storage.Appendable] with an implementation that can be
// tokenized as a specific string.
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
	// from the calling converter code.
	if len(f.Targets) > 1 {
		toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "concat"})
		toks = append(toks, builder.Token{Tok: token.LPAREN, Lit: ""})
	}

	for ix, targetMap := range f.Targets {
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
