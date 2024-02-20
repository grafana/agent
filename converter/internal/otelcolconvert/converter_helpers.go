package otelcolconvert

import (
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/river/token"
	"github.com/grafana/river/token/builder"
)

// This file contains shared helpers for converters to use.

// tokenizedConsumer implements [otelcol.Consumer] and [builder.Tokenizer].
// tokenizedConsumer tokenizes as the string literal specified by the Expr
// field.
type tokenizedConsumer struct {
	otelcol.Consumer

	Expr string // Expr is the string to return during tokenization.
}

func (tc tokenizedConsumer) RiverCapsule() {}

func (tc tokenizedConsumer) RiverTokenize() []builder.Token {
	return []builder.Token{{
		Tok: token.STRING,
		Lit: tc.Expr,
	}}
}
