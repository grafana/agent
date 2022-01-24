package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_jsonnetMarshal(t *testing.T) {
	tt := []struct {
		name   string
		in     interface{}
		expect string
	}{
		{name: "string", in: "hello", expect: `"hello"`},
		{name: "bool", in: true, expect: `true`},
		{name: "number", in: 5, expect: `5`},
		{name: "array", in: []int{0, 1, 2}, expect: `[0, 1, 2]`},

		{
			name: "struct",
			in: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{"John", 42},
			expect: `{"Name":"John","Age":42}`,
		},
		{
			name: "array of structs",
			in: func() interface{} {
				type ty struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}
				return []ty{{"John", 42}, {"Anne", 41}, {"Peter", 40}}
			}(),
			expect: `[
        {"Name": "John", "Age": 42},
        {"Name": "Anne", "Age": 41},
        {"Name": "Peter", "Age": 40}
      ]`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := jsonnetMarshal(tc.in)
			require.NoError(t, err)
			require.JSONEq(t, tc.expect, string(actual))
		})
	}
}
