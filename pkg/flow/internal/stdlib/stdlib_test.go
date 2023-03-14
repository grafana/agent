package stdlib

import (
	"reflect"
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestVM_Stdlib_Scoped(t *testing.T) {
	rootScope := &vm.Scope{
		Variables: Identifiers,
	}

	tt := []struct {
		name   string
		input  string
		scope  *vm.Scope
		expect interface{}
	}{
		{
			name:  "discovery_target_decode",
			input: `discovery_target_decode(input)`,
			scope: &vm.Scope{
				Parent: rootScope,
				Variables: map[string]interface{}{
					"input": `[
						{
							"targets": ["host-a:12345", "host-a:12346"],
							"labels": {
								"foo": "bar"
							}
						},
						{
							"targets": ["host-b:12345", "host-b:12346"],
							"labels": {
								"hello": "world"
							}
						}
					]`,
				},
			},
			expect: []discovery.Target{
				{
					model.AddressLabel: "host-a:12345",
					"foo":              "bar",
				},
				{
					model.AddressLabel: "host-a:12346",
					"foo":              "bar",
				},
				{
					model.AddressLabel: "host-b:12345",
					"hello":            "world",
				},
				{
					model.AddressLabel: "host-b:12346",
					"hello":            "world",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tc.input)
			require.NoError(t, err)

			eval := vm.New(expr)

			rv := reflect.New(reflect.TypeOf(tc.expect))
			require.NoError(t, eval.Evaluate(tc.scope, rv.Interface()))
			require.Equal(t, tc.expect, rv.Elem().Interface())
		})
	}
}
