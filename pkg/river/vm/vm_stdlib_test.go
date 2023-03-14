package vm_test

import (
	"fmt"
	"reflect"
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
		{"json_decode object", `json_decode("{\"foo\": \"bar\"}")`, map[string]interface{}{"foo": "bar"}},
		{"json_decode array", `json_decode("[0, 1, 2]")`, []interface{}{float64(0), float64(1), float64(2)}},
		{"json_decode nil field", `json_decode("{\"foo\": null}")`, map[string]interface{}{"foo": nil}},
		{"json_decode nil array element", `json_decode("[0, null]")`, []interface{}{float64(0), nil}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tc.input)
			require.NoError(t, err)

			eval := vm.New(expr)

			rv := reflect.New(reflect.TypeOf(tc.expect))
			require.NoError(t, eval.Evaluate(nil, rv.Interface()))
			require.Equal(t, tc.expect, rv.Elem().Interface())
		})
	}
}

func BenchmarkConcat(b *testing.B) {
	// There's a bit of setup work to do here: we want to create a scope holding
	// a slice of the Person type, which has a fair amount of data in it.
	//
	// We then want to pass it through concat.
	//
	// If the code path is fully optimized, there will be no intermediate
	// translations to interface{}.
	type Person struct {
		Name  string            `river:"name,attr"`
		Attrs map[string]string `river:"attrs,attr"`
	}
	type Body struct {
		Values []Person `river:"values,attr"`
	}

	in := `values = concat(values_ref)`
	f, err := parser.ParseFile("", []byte(in))
	require.NoError(b, err)

	eval := vm.New(f)

	valuesRef := make([]Person, 0, 20)
	for i := 0; i < 20; i++ {
		data := make(map[string]string, 20)
		for j := 0; j < 20; j++ {
			var (
				key   = fmt.Sprintf("key_%d", i+1)
				value = fmt.Sprintf("value_%d", i+1)
			)
			data[key] = value
		}
		valuesRef = append(valuesRef, Person{
			Name:  "Test Person",
			Attrs: data,
		})
	}
	scope := &vm.Scope{
		Variables: map[string]interface{}{
			"values_ref": valuesRef,
		},
	}

	// Reset timer before running the actual test
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var b Body
		_ = eval.Evaluate(scope, &b)
	}
}
