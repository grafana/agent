package common

import (
	"github.com/grafana/agent/service/labelstore"
	"github.com/grafana/river"
	"github.com/grafana/river/token"
	"github.com/grafana/river/token/builder"
)

// ConvertAppendable implements both the [builder.Tokenizer] and
// [storage.Appendable] interfaces. This allows us to set component.Arguments
// that leverage [storage.Appendable] with an implementation that can be
// tokenized as a specific string.
type ConvertAppendable struct {
	labelstore.Appendable

	Expr string // The specific string to return during tokenization.
}

var (
	_ labelstore.Appendable = (*ConvertAppendable)(nil)
	_ builder.Tokenizer     = ConvertAppendable{}
	_ river.Capsule         = ConvertAppendable{}
)

func (f ConvertAppendable) RiverCapsule() {}
func (f ConvertAppendable) RiverTokenize() []builder.Token {
	return []builder.Token{{
		Tok: token.STRING,
		Lit: f.Expr,
	}}
}
