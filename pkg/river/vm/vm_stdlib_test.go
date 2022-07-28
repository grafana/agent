package vm_test

import (
	"testing"

	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/stretchr/testify/require"
)

func TestVM_Stdlib(t *testing.T) {
	t.Setenv("TEST_VAR", "Hello!")

	tt := []struct {
		name   string
		input  string
		expect interface{}
	}{
		{"env", `env("TEST_VAR")`, string("Hello!")},
		{"concat", `concat([true, "foo"], [], [false, 1])`, []interface{}{true, "foo", false, 1}},
		{"unmarshal_json object", `unmarshal_json("{\"foo\": \"bar\"}")`, map[string]interface{}{"foo": "bar"}},
		{"unmarshal_json array", `unmarshal_json("[0, 1, 2]")`, []interface{}{float64(0), float64(1), float64(2)}},
		{"unmarshal_json nil field", `unmarshal_json("{\"foo\": null}")`, map[string]interface{}{"foo": nil}},
		{"unmarshal_json nil array element", `unmarshal_json("[0, null]")`, []interface{}{float64(0), nil}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tc.input)
			require.NoError(t, err)

			eval := vm.New(expr)

			var v interface{}
			require.NoError(t, eval.Evaluate(nil, &v))
			require.Equal(t, tc.expect, v)
		})
	}
}
