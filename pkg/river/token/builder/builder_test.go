package builder_test

import (
	"bytes"
	"testing"

	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/stretchr/testify/require"
)

func TestBuilder_File(t *testing.T) {
	f := builder.NewFile()

	f.Body().SetAttributeTokens("attr_1", []builder.Token{{Tok: token.NUMBER, Lit: "15"}})
	f.Body().SetAttributeTokens("attr_2", []builder.Token{{Tok: token.BOOL, Lit: "true"}})

	b1 := builder.NewBlock([]string{"test", "block"}, "")
	b1.Body().SetAttributeTokens("inner_attr", []builder.Token{{Tok: token.STRING, Lit: `"block 1"`}})
	f.Body().AppendBlock(b1)

	b2 := builder.NewBlock([]string{"test", "block"}, "labeled")
	b2.Body().SetAttributeTokens("inner_attr", []builder.Token{{Tok: token.STRING, Lit: `"block 2"`}})
	f.Body().AppendBlock(b2)

	expect := format(t, `
		attr_1 = 15
		attr_2 = true

		test.block {
			inner_attr = "block 1"
		}

		test.block "labeled" {
			inner_attr = "block 2"
		}
	`)

	require.Equal(t, expect, string(f.Bytes()))
}

func TestBuilder_GoEncode(t *testing.T) {
	f := builder.NewFile()

	f.Body().SetAttributeValue("null_value", nil)
	f.Body().AppendTokens([]builder.Token{{token.LITERAL, "\n"}})

	f.Body().SetAttributeValue("num", 15)
	f.Body().SetAttributeValue("string", "Hello, world!")
	f.Body().SetAttributeValue("bool", true)
	f.Body().SetAttributeValue("list", []int{0, 1, 2})
	f.Body().SetAttributeValue("func", func(int, int) int { return 0 })
	f.Body().SetAttributeValue("capsule", make(chan int))
	f.Body().AppendTokens([]builder.Token{{token.LITERAL, "\n"}})

	f.Body().SetAttributeValue("map", map[string]interface{}{"foo": "bar"})
	f.Body().SetAttributeValue("map_2", map[string]interface{}{"non ident": "bar"})
	f.Body().AppendTokens([]builder.Token{{token.LITERAL, "\n"}})

	f.Body().SetAttributeValue("mixed_list", []interface{}{
		0,
		true,
		map[string]interface{}{"key": true},
		"Hello!",
	})

	expect := format(t, `
		null_value = null
	
		num     = 15 
		string  = "Hello, world!"
		bool    = true
		list    = [0, 1, 2]
		func    = function
		capsule = capsule("chan int")

		map = {
			foo = "bar",
		}
		map_2 = {
			"non ident" = "bar",
		}

		mixed_list = [0, true, {
			key = true,
		}, "Hello!"]
	`)

	require.Equal(t, expect, string(f.Bytes()))
}

func format(t *testing.T, in string) string {
	t.Helper()

	f, err := parser.ParseFile(t.Name(), []byte(in))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, printer.Fprint(&buf, f))

	return buf.String()
}

type CustomTokenizer bool

var _ builder.Tokenizer = (CustomTokenizer)(false)

func (ct CustomTokenizer) RiverTokenize() []builder.Token {
	return []builder.Token{{Tok: token.LITERAL, Lit: "CUSTOM_TOKENS"}}
}

func TestBuilder_GoEncode_Tokenizer(t *testing.T) {
	f := builder.NewFile()

	f.Body().SetAttributeValue("custom_tokens", map[string]interface{}{
		"number":           15,
		"custom_tokenizer": CustomTokenizer(true),
		"string":           "Hello, world!",
	})

	expect := format(t, `
		custom_tokens = {
			number = 15,
			custom_tokenizer = CUSTOM_TOKENS,
			string = "Hello, world!",
		}
	`)

	require.Equal(t, expect, string(f.Bytes()))
}
