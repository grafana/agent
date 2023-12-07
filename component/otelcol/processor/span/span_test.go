package span_test

import (
	"context"
	"testing"

	"github.com/grafana/agent/component/otelcol/processor/processortest"
	"github.com/grafana/agent/component/otelcol/processor/span"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	tests := []struct {
		flowCfg              string
		otelCfg              map[string]interface{}
		expectUnmarshalError bool
	}{
		{
			flowCfg: `
			name {
				separator    = "::"
				from_attributes  = ["db.svc", "operation", "id"]
			}

			output {}
			`,
			otelCfg: map[string]interface{}{
				"name": spanprocessor.Name{
					FromAttributes: []string{"db.svc", "operation", "id"},
					Separator:      "::",
				},
			},
		},
		{
			flowCfg: `
			name {
				from_attributes  = ["db.svc", "operation", "id"]
			}

			output {}
			`,
			otelCfg: map[string]interface{}{
				"name": spanprocessor.Name{
					FromAttributes: []string{"db.svc", "operation", "id"},
				},
			},
		},
		{
			flowCfg: `
			name {
				to_attributes {
					rules = ["^\\/api\\/v1\\/document\\/(?P<documentId>.*)\\/update$"]
				}
			}

			output {}
			`,
			otelCfg: map[string]interface{}{
				"name": spanprocessor.Name{
					ToAttributes: &spanprocessor.ToAttributes{
						Rules: []string{`^\/api\/v1\/document\/(?P<documentId>.*)\/update$`},
					},
				},
			},
		},
		{
			flowCfg: `
			include {
				match_type = "regexp"
				services   = ["banks"]
				span_names = ["^(.*?)/(.*?)$"]
			}
			exclude {
				match_type = "strict"
				span_names = ["donot/change"]
			}
			name {
				to_attributes {
					rules  = ["(?P<operation_website>.*?)$"]
				}
			}

			output {}
			`,
			otelCfg: map[string]interface{}{
				"include": map[string]interface{}{
					"match_type": "regexp",
					"services":   []string{"banks"},
					"span_names": []string{`^(.*?)/(.*?)$`},
				},
				"exclude": map[string]interface{}{
					"match_type": "strict",
					"span_names": []string{`donot/change`},
				},
				"name": spanprocessor.Name{
					ToAttributes: &spanprocessor.ToAttributes{
						Rules: []string{`(?P<operation_website>.*?)$`},
					},
				},
			},
		},
		{
			flowCfg: `
			status {
				code  =  "Error"
				description = "some additional error description"
			}

			output {}
			`,
			otelCfg: map[string]interface{}{
				"status": spanprocessor.Status{
					Code:        "Error",
					Description: "some additional error description",
				},
			},
		},
		{
			flowCfg: `
			include {
				match_type = "strict"
				attribute {
					key = "http.status_code"
					value = 400
				}
			}
			status {
				code  =  "Ok"
			}

			output {}
			`,
			otelCfg: map[string]interface{}{
				"include": map[string]interface{}{
					"match_type": "strict",
					"attributes": []interface{}{
						map[string]interface{}{
							"key":   "http.status_code",
							"value": 400,
						},
					},
				},
				"status": spanprocessor.Status{
					Code: "Ok",
				},
			},
		},
	}

	for _, tc := range tests {
		var args span.Arguments
		err := river.Unmarshal([]byte(tc.flowCfg), &args)

		if tc.expectUnmarshalError {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)

		ext, err := args.Convert()

		require.NoError(t, err)
		otelArgs, ok := (ext).(*spanprocessor.Config)
		require.True(t, ok)

		var expectedArgs spanprocessor.Config
		require.NoError(t, mapstructure.Decode(tc.otelCfg, &expectedArgs))

		require.Equal(t, expectedArgs, *otelArgs)
	}
}

// Below are tests which run a whole processor from end to end.
// Their configs are inspired by the example configs in the otel repo:
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/spanprocessor/testdata/config.yaml

func testRunProcessor(t *testing.T, processorConfig string, testSignal processortest.Signal) {
	ctx := componenttest.TestContext(t)
	testRunProcessorWithContext(ctx, t, processorConfig, testSignal)
}

func testRunProcessorWithContext(ctx context.Context, t *testing.T, processorConfig string, testSignal processortest.Signal) {
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.span")
	require.NoError(t, err)

	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(processorConfig), &args))

	// Override the arguments so signals get forwarded to the test channel.
	args.Output = testSignal.MakeOutput()

	prc := processortest.ProcessorRunConfig{
		Ctx:        ctx,
		T:          t,
		Args:       args,
		TestSignal: testSignal,
		Ctrl:       ctrl,
		L:          l,
	}
	processortest.TestRunProcessor(prc)
}

func Test_UpdateSpanNameFromAttributesSuccessfully(t *testing.T) {
	cfg := `
	name {
		separator    = "::"
		from_attributes  = ["db.svc", "operation", "id"]
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, "::", otelArgs.Rename.Separator)
	require.Equal(t, 3, len(otelArgs.Rename.FromAttributes))
	require.Equal(t, "db.svc", otelArgs.Rename.FromAttributes[0])
	require.Equal(t, "operation", otelArgs.Rename.FromAttributes[1])
	require.Equal(t, "id", otelArgs.Rename.FromAttributes[2])

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "serviceA",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "operation",
						"value": { "stringValue": "get" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "location::get::1234",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "operation",
						"value": { "stringValue": "get" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_UpdateSpanNameFromAttributesUnsuccessfully(t *testing.T) {
	cfg := `
	name {
		separator    = "::"
		from_attributes  = ["db.svc", "operation", "id"]
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, "::", otelArgs.Rename.Separator)
	require.Equal(t, 3, len(otelArgs.Rename.FromAttributes))
	require.Equal(t, "db.svc", otelArgs.Rename.FromAttributes[0])
	require.Equal(t, "operation", otelArgs.Rename.FromAttributes[1])
	require.Equal(t, "id", otelArgs.Rename.FromAttributes[2])

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "serviceA",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "serviceA",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_UpdateSpanNameFromAttributesNoSeparatorSuccessfully(t *testing.T) {
	cfg := `
	name {
		from_attributes  = ["db.svc", "operation", "id"]
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, 3, len(otelArgs.Rename.FromAttributes))
	require.Equal(t, "db.svc", otelArgs.Rename.FromAttributes[0])
	require.Equal(t, "operation", otelArgs.Rename.FromAttributes[1])
	require.Equal(t, "id", otelArgs.Rename.FromAttributes[2])

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "serviceA",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "operation",
						"value": { "stringValue": "get" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "locationget1234",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "operation",
						"value": { "stringValue": "get" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ToAttributes(t *testing.T) {
	cfg := `
	name {
		to_attributes {
			rules = ["^\\/api\\/v1\\/document\\/(?P<documentId>.*)\\/update$"]
		}
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, 1, len(otelArgs.Rename.ToAttributes.Rules))
	require.Equal(t, `^\/api\/v1\/document\/(?P<documentId>.*)\/update$`, otelArgs.Rename.ToAttributes.Rules[0])

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "/api/v1/document/12345678/update",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "operation",
						"value": { "stringValue": "get" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "/api/v1/document/{documentId}/update",
					"attributes": [{
						"key": "db.svc",
						"value": { "stringValue": "location" }
					},{
						"key": "operation",
						"value": { "stringValue": "get" }
					},{
						"key": "id",
						"value": { "intValue": "1234" }
					},{
						"key": "documentId",
						"value": { "stringValue": "12345678" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_IncludeExclude(t *testing.T) {
	cfg := `
	include {
		match_type = "regexp"
		services   = ["banks"]
		span_names = ["^(.*?)/(.*?)$"]
	}
	exclude {
		match_type = "strict"
		span_names = ["donot/change"]
	}
	name {
		to_attributes {
			rules  = ["(?P<operation_website>.*?)$"]
		}
	}

	output {
		// no-op: will be overridden by test code.
	}
`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, 1, len(otelArgs.Rename.ToAttributes.Rules))
	require.Equal(t, `(?P<operation_website>.*?)$`, otelArgs.Rename.ToAttributes.Rules[0])

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "banks" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "Span/1",
					"attributes": []
				}]
			}]
		},{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "SvcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "Span/1",
					"attributes": []
				}]
			}]
		},{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "banks" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "donot/change",
					"attributes": []
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "banks" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "{operation_website}",
					"attributes": [{
						"key": "operation_website",
						"value": { "stringValue": "Span/1" }
					}]
				}]
			}]
		},{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "SvcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "Span/1",
					"attributes": []
				}]
			}]
		},{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "banks" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "donot/change",
					"attributes": []
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_StatusError(t *testing.T) {
	cfg := `
	status {
		code  =  "Error"
		description = "some additional error description"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, "Error", otelArgs.SetStatus.Code)
	require.Equal(t, "some additional error description", otelArgs.SetStatus.Description)

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {},
			"scopeSpans": [{
				"scope": {},
				"spans": [{
					"name": "TestSpan",
					"status": {}
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {},
			"scopeSpans": [{
				"scope": {},
				"spans": [{
					"name": "TestSpan",
					"status": {
						"code":2, 
						"message":"some additional error description"
					}
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_StatusOk(t *testing.T) {
	cfg := `
	include {
		match_type = "strict"
		attribute {
			key = "http.status_code"
			value = 400
		}
	}
	status {
		code  =  "Ok"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args span.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (convertedArgs).(*spanprocessor.Config)
	require.True(t, ok)

	require.Equal(t, "Ok", otelArgs.SetStatus.Code)

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "http.status_code",
						"value": { "intValue": "400" }
					}],
					"status": {}
				}]
			}]
		},{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpa2",
					"attributes": [{
						"key": "http.status_code",
						"value": { "intValue": "500" }
					}],
					"status": {}
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "http.status_code",
						"value": { "intValue": "400" }
					}],
					"status": {
						"code":1
					}
				}]
			}]
		},{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpa2",
					"attributes": [{
						"key": "http.status_code",
						"value": { "intValue": "500" }
					}],
					"status": {}
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTrace, expectedOutputTrace))
}
