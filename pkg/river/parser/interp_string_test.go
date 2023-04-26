package parser

import (
	"testing"

	"github.com/grafana/agent/pkg/river/token"
	"github.com/stretchr/testify/assert"
)

type result struct {
	Pos int
	Tok interpStringToken
	Lit string
}

func mergeResults(results []result) []result {
	var out []result

	for i, result := range results {
		// Merge results if they're successive sequences of interpStringTokenRaw.
		if i > 0 && result.Tok == interpStringTokenRaw && out[len(out)-1].Tok == interpStringTokenRaw {
			out[len(out)-1].Lit += result.Lit
			continue
		}

		out = append(out, result)
	}

	return out
}

func Test_interpStringScanner(t *testing.T) {
	tt := []struct {
		input  string
		expect []result
	}{
		{
			input:  "",
			expect: nil,
		},
		{
			input: "hello, world",
			expect: []result{
				{0, interpStringTokenRaw, "hello, world"},
			},
		},
		{
			input: "${1+2}",
			expect: []result{
				{0, interpStringTokenExpr, "${1+2}"},
			},
		},
		{
			// Escaped sequence.
			input: "\\${1+2}",
			expect: []result{
				{0, interpStringTokenRaw, "\\${1+2}"},
			},
		},
		{
			// Unterminated interpolated expression.
			input: "Hello, world! ${1+2",
			expect: []result{
				{0, interpStringTokenRaw, "Hello, world! ${1+2"},
			},
		},
		{
			// Multiple pairs of curly braces.
			input: "${{ a = 5 }}",
			expect: []result{
				{0, interpStringTokenExpr, "${{ a = 5 }}"},
			},
		},
		{
			input: "The Matrix came out in ${1999}.",
			expect: []result{
				{0, interpStringTokenRaw, "The Matrix came out in "},
				{23, interpStringTokenExpr, "${1999}"},
				{30, interpStringTokenRaw, "."},
			},
		},
	}

	for _, tc := range tt {
		f := token.NewFile(t.Name())
		s := newInterpStringScanner(f, []byte(tc.input), nil, 0)

		var results []result
		for {
			pos, tok, lit := s.Scan()
			if tok == interpStringTokenEOF {
				break
			}

			results = append(results, result{
				Pos: pos.Offset(),
				Tok: tok,
				Lit: lit,
			})
		}

		results = mergeResults(results)
		assert.Equal(t, tc.expect, results, "Unexpected result for scanning %q", tc.input)
	}
}
