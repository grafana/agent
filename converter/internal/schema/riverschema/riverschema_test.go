package riverschema_test

import (
	"bytes"
	"fmt"
	"testing"

	schema "github.com/grafana/agent/converter/internal/schema/riverschema"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/stretchr/testify/assert"
)

// Example demonstrates what a generated schema for a component may look like.
// otelcol.receiver.opencensus is used as an example.
func Example() {
	type (
		ConsumerArguments struct {
			Metrics *schema.Array[*schema.Capsule] `river:"metrics,attr,optional"`
			Logs    *schema.Array[*schema.Capsule] `river:"logs,attr,optional"`
			Traces  *schema.Array[*schema.Capsule] `river:"traces,attr,optional"`
		}

		GRPCServerArguments struct {
			Endpoint *schema.String `river:"endpoint,attr,optional"`
			// (other keys are omitted for brevity)
		}

		Arguments struct {
			CorsAllowedOrigins *schema.Array[*schema.String] `river:"cors_allowed_origins,attr,optional"`
			GRPC               *GRPCServerArguments          `river:"grpc,block,optional"`
			Output             *ConsumerArguments            `river:"output,block"`
		}
	)

	// Create our component arguments.
	component := &Arguments{
		GRPC: &GRPCServerArguments{
			Endpoint: schema.NewString("0.0.0.0:4317"),
		},
		Output: &ConsumerArguments{
			Metrics: schema.NewArray([]*schema.Capsule{
				schema.ExprCapsule("otelcol.exporter.otlp.default.input"),
			}),
		},
	}

	// Tokenize our component arguments out.
	f := builder.NewFile()
	b := builder.NewBlock([]string{"otelcol", "receiver", "opencensus"}, "example")
	b.Body().AppendFrom(component)
	f.Body().AppendBlock(b)
	fmt.Println(string(f.Bytes()))
	// OUTPUT: otelcol.receiver.opencensus "example" {
	// 	grpc {
	//		endpoint = "0.0.0.0:4317"
	// 	}
	//
	// 	output {
	// 		metrics = [otelcol.exporter.otlp.default.input]
	// 	}
	// }
}

func Test(t *testing.T) {
	tt := []struct {
		input  schema.Type
		expect string
	}{
		{schema.NewUnsignedNumber(15), `15`},
		{schema.NewSignedNumber(-15), `-15`},
		{schema.NewFloatNumber(1.2345), `1.2345`},
		{schema.NewString("Hello, world!"), `"Hello, world!"`},
		{schema.NewBool(true), `true`},
		{schema.NewArray[schema.Type](nil), `[]`},
		{
			input: schema.NewArray([]schema.Type{
				schema.NewBool(true),
				schema.NewFloatNumber(1.2345),
				schema.NewString("Hello, world!"),
			}),
			expect: `[true, 1.2345, "Hello, world!"]`,
		},
		{schema.NewObject[schema.Type](nil), `{}`},
		{
			input: schema.NewObject(map[string]schema.Type{
				"key_a": schema.NewBool(true),
				"key_b": schema.NewFloatNumber(1.2345),
				"key_c": schema.NewString("Hello, world!"),
			}),
			expect: `{
	"key_a" = true,
	"key_b" = 1.2345,
	"key_c" = "Hello, world!",
}`,
		},

		{schema.ExprUnsignedNumber(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprSignedNumber(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprFloatNumber(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprString(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprBool(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprArray[schema.Type](`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprObject[schema.Type](`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprFunction(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
		{schema.ExprCapsule(`env("TEST_EXPR")`), `env("TEST_EXPR")`},
	}

	for _, tc := range tt {
		expr := builder.NewExpr()
		expr.SetValue(tc.input)

		var buf bytes.Buffer
		_, err := expr.WriteTo(&buf)
		if !assert.NoError(t, err, "%#v did not tokenize properly", tc.input) {
			continue
		}

		assert.Equal(t, tc.expect, string(expr.Bytes()), "%#v did not print to expected output", tc.input)
	}
}
