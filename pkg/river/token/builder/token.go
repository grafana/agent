package builder

import (
	"bytes"
	"io"

	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
	"github.com/grafana/agent/pkg/river/token"
)

// A Token is a wrapper around token.Token which contains the token type
// alongside its literal. Use token.ILLEGAL to write literal characters such as
// whitespace.
type Token struct {
	Type token.Token
	Lit  string
}

// printTokens prints out the tokens as River text and formats them, writing
// the final result to w.
func printTokens(w io.Writer, toks []Token) (int, error) {
	var raw bytes.Buffer
	for _, tok := range toks {
		switch {
		case tok.Type == token.ILLEGAL:
			raw.WriteString(tok.Lit)
		case tok.Type.IsLiteral() || tok.Type.IsKeyword():
			raw.WriteString(tok.Lit)
		default:
			raw.WriteString(tok.Type.String())
		}
	}

	f, err := parser.ParseFile("", raw.Bytes())
	if err != nil {
		return 0, err
	}

	wc := &writerCount{w: w}
	err = printer.Fprint(wc, f)
	return wc.n, err
}

type writerCount struct {
	w io.Writer
	n int
}

func (wc *writerCount) Write(p []byte) (n int, err error) {
	n, err = wc.w.Write(p)
	wc.n += n
	return
}
