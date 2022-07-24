package builder

import (
	"encoding"
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/grafana/agent/pkg/river/scanner"
	"github.com/grafana/agent/pkg/river/token"
)

// TODO(rfratto): check for optional values

// Tokenizer is any value which can return a raw set of tokens.
type Tokenizer interface {
	// RiverTokenize returns the raw set of River tokens which are used when
	// printing out the value with river/token/builder.
	RiverTokenize() []Token
}

func tokenEncode(val interface{}) []Token {
	return valueTokens(value.Encode(val))
}

func valueTokens(v value.Value) []Token {
	var toks []Token

	// Check for interfces which override encoding behavior:
	switch v := v.Interface().(type) {
	case time.Duration:
		return []Token{{
			Tok: token.STRING,
			Lit: fmt.Sprintf("%q", v.String()),
		}}
	case Tokenizer:
		return v.RiverTokenize()
	case encoding.TextMarshaler:
		// TODO(rfratto): what if this fails?
		bb, _ := v.MarshalText()
		return []Token{{
			Tok: token.STRING,
			Lit: fmt.Sprintf("%q", string(bb)),
		}}
	}

	switch v.Type() {
	case value.TypeNull:
		toks = append(toks, Token{token.NULL, "null"})

	case value.TypeNumber:
		toks = append(toks, Token{token.NUMBER, v.Number().ToString()})

	case value.TypeString:
		toks = append(toks, Token{token.STRING, fmt.Sprintf("%q", v.Text())})

	case value.TypeBool:
		toks = append(toks, Token{token.STRING, fmt.Sprintf("%v", v.Bool())})

	case value.TypeArray:
		toks = append(toks, Token{token.LBRACK, ""})
		elems := v.Len()
		for i := 0; i < elems; i++ {
			elem := v.Index(i)

			toks = append(toks, valueTokens(elem)...)
			if i+1 < elems {
				toks = append(toks, Token{token.COMMA, ""})
			}
		}
		toks = append(toks, Token{token.RBRACK, ""})

	case value.TypeObject:
		toks = append(toks, Token{token.LCURLY, ""}, Token{token.LITERAL, "\n"})

		keys := v.Keys()
		for i := 0; i < len(keys); i++ {
			if isValidIdentifier(keys[i]) {
				toks = append(toks, Token{token.IDENT, keys[i]})
			} else {
				toks = append(toks, Token{token.STRING, fmt.Sprintf("%q", keys[i])})
			}

			field, _ := v.Key(keys[i])
			toks = append(toks, Token{token.ASSIGN, ""})
			toks = append(toks, valueTokens(field)...)
			toks = append(toks, Token{token.COMMA, ""}, Token{token.LITERAL, "\n"})
		}
		toks = append(toks, Token{token.RCURLY, ""})

	case value.TypeFunction:
		toks = append(toks, Token{token.LITERAL, v.Describe()})

	case value.TypeCapsule:
		toks = append(toks, Token{token.LITERAL, v.Describe()})

	default:
		panic(fmt.Sprintf("river/token/builder: unrecognized value type %q", v.Type()))
	}

	return toks
}

func isValidIdentifier(in string) bool {
	s := scanner.New(nil, []byte(in), nil, 0)
	_, tok, lit := s.Scan()
	return tok == token.IDENT && lit == in
}
