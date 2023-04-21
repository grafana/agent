package attributes_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/component/otelcol/processor/attributes"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/dskit/backoff"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// A lot of the TestDecode tests were inspired by tests in the Otel repo:
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.63.0/processor/attributesprocessor/testdata/config.yaml

func Test_Insert(t *testing.T) {
	cfg := `
		action {
			key = "attribute1"
			value = 111111
			action = "insert"
		}
		action {
			key = "string key"
			value = "anotherkey"
			action = "insert"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "attribute1", action.Key)
	require.Equal(t, 111111, action.Value)
	require.Equal(t, "insert", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "string key", action.Key)
	require.Equal(t, "anotherkey", action.Value)
	require.Equal(t, "insert", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "attribute1",
						"value": { "intValue": "0" }
					}]
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
						"traceId": "",
						"spanId": "",
						"parentSpanId": "",
						"name": "TestSpan",
						"attributes": [{
							"key": "attribute1",
							"value": { "intValue": "0" }
						},
						{
							"key": "string key",
							"value": { "stringValue": "anotherkey" }
						}],
						"status": {}
					}]
				}]
			}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_RegexExtract(t *testing.T) {
	// cfg := `
	// 	action {
	// 		key = "http.url"
	// 		pattern = "^(?P<http_protocol>.*):\\/\\/(?P<http_domain>.*)\\/(?P<http_path>.*)(\\?|\\&)(?P<http_query_params>.*)"
	// 		action = "extract"
	// 	}

	// 	output {
	// 		// no-op: will be overridden by test code.
	// 	}
	// `
	cfg := `
		action {
			key = "http.url"
			pattern = "^(?P<http_protocol>.*):\\/\\/(?P<http_domain>.*)\\/(?P<http_path>.*)(\\?|\\&)(?P<http_query_params>.*)"
			action = "extract"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "http.url", action.Key)
	// require.Equal(t, "^(?P<http_protocol>.*):\\/\\/(?P<http_domain>.*)\\/(?P<http_path>.*)(\\?|\\&)(?P<http_query_params>.*)", action.RegexPattern)
	require.Equal(t, "extract", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "http.url",
						"value": { "stringValue": "http://example.com/path?queryParam1=value1,queryParam2=value2" }
					}]
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
						"traceId": "",
						"spanId": "",
						"parentSpanId": "",
						"name": "TestSpan",
						"attributes": [{
							"key": "attribute1",
							"value": { "intValue": "0" }
						},
						{
							"key": "string key",
							"value": { "stringValue": "anotherkey" }
						}],
						"status": {}
					}]
				}]
			}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Update(t *testing.T) {
	cfg := `
		action {
			key = "boo"
			from_attribute = "foo"
			action = "update"
		}
		action {
			key = "db.secret"
			value = "redacted"
			action = "update"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "boo", action.Key)
	require.Equal(t, "foo", action.FromAttribute)
	require.Equal(t, "update", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "db.secret", action.Key)
	require.Equal(t, "redacted", action.Value)
	require.Equal(t, "update", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "foo",
						"value": { "intValue": "11111" }
					},
					{
						"key": "boo",
						"value": { "intValue": "22222" }
					},
					{
						"key": "db.secret",
						"value": { "stringValue": "top_secret" }
					}]
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
						"key": "foo",
						"value": { "intValue": "11111" }
					},
					{
						"key": "boo",
						"value": { "intValue": "11111" }
					},
					{
						"key": "db.secret",
						"value": { "stringValue": "redacted" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Upsert(t *testing.T) {
	cfg := `
		action {
			key = "region"
			value = "planet-earth"
			action = "upsert"
		}
		action {
			key = "new_user_key"
			from_attribute = "user_key"
			action = "upsert"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "region", action.Key)
	require.Equal(t, "planet-earth", action.Value)
	require.Equal(t, "upsert", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "new_user_key", action.Key)
	require.Equal(t, "user_key", action.FromAttribute)
	require.Equal(t, "upsert", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "user_key",
						"value": { "intValue": "11111" }
					}]
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
						"key": "user_key",
						"value": { "intValue": "11111" }
					},
					{
						"key": "region",
						"value": { "stringValue": "planet-earth" }
					},
					{
						"key": "new_user_key",
						"value": { "intValue": "11111" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Delete(t *testing.T) {
	cfg := `
		action {
			key = "credit_card"
			action = "delete"
		}
		action {
			key = "duplicate_key"
			action = "delete"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "credit_card",
						"value": { "intValue": "11111" }
					},
					{
						"key": "duplicate_key",
						"value": { "intValue": "22222" }
					},
					{
						"key": "db.secret",
						"value": { "stringValue": "top_secret" }
					}]
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
						"key": "db.secret",
						"value": { "stringValue": "top_secret" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Hash(t *testing.T) {
	cfg := `
		action {
			key = "user.email"
			action = "hash"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "user.email", action.Key)
	require.Equal(t, "hash", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "foo",
						"value": { "intValue": "11111" }
					},
					{
						"key": "boo",
						"value": { "intValue": "22222" }
					},
					{
						"key": "user.email",
						"value": { "stringValue": "user@email.com" }
					}]
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
						"key": "foo",
						"value": { "intValue": "11111" }
					},
					{
						"key": "boo",
						"value": { "intValue": "22222" }
					},
					{
						"key": "user.email",
						"value": { "stringValue": "36687c352204c27d9e228a9b34d00c8a1d36a000" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Convert(t *testing.T) {
	cfg := `
		action {
			key = "http.status_code"
			converted_type = "int"
			action = "convert"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "http.status_code", action.Key)
	require.Equal(t, "int", action.ConvertedType)
	require.Equal(t, "convert", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan",
					"attributes": [{
						"key": "http.status_code",
						"value": { "stringValue": "500" }
					}]
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
						"value": { "intValue": "500" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExcludeMulti(t *testing.T) {
	cfg := `
	exclude {
		match_type = "strict"
		services = ["svcA", "svcB"]
		attribute {
			key = "env"
			value = "dev"
		}
		attribute {
			key = "test_request"
		}
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Exclude)

	require.Equal(t, "strict", string(otelObj.Exclude.MatchType))

	svc := &otelObj.Exclude.Services[0]
	require.Equal(t, "svcA", *svc)
	svc = &otelObj.Exclude.Services[1]
	require.Equal(t, "svcB", *svc)

	attr := &otelObj.Exclude.Attributes[0]
	require.Equal(t, "env", attr.Key)
	require.Equal(t, "dev", attr.Value)

	attr = &otelObj.Exclude.Attributes[1]
	require.Equal(t, "test_request", attr.Key)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcC" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcC" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExcludeResources(t *testing.T) {
	cfg := `
	exclude {
		match_type = "strict"
		resource {
			key = "host.type"
			value = "n1-standard-1"
		}
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Exclude)

	res := &otelObj.Exclude.Resources[0]
	require.Equal(t, "strict", string(otelObj.Exclude.MatchType))

	require.Equal(t, "host.type", res.Key)
	require.Equal(t, "n1-standard-1", res.Value)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExcludeLibrary(t *testing.T) {
	cfg := `
	exclude {
		match_type = "strict"
		library {
			name = "mongo-java-driver"
			version = "3.8.0"
		}
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Exclude)
	require.Equal(t, "strict", string(otelObj.Exclude.MatchType))

	lib := &otelObj.Exclude.Libraries[0]
	require.Equal(t, "mongo-java-driver", lib.Name)
	require.Equal(t, "3.8.0", *lib.Version)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": "3.8.0"
				},
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "dummy-driver",
					"version": "1.1.0"
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": "3.8.0"
				},
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "dummy-driver",
					"version": "1.1.0"
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExcludeLibraryAnyVersion(t *testing.T) {
	cfg := `
	exclude {
		match_type = "strict"
		library {
			name = "mongo-java-driver"
		}
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Exclude)
	require.Equal(t, "strict", string(otelObj.Exclude.MatchType))

	lib := &otelObj.Exclude.Libraries[0]
	require.Equal(t, "mongo-java-driver", lib.Name)
	require.Nil(t, lib.Version)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": "3.8.0"
				},
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "dummy-driver",
					"version": "1.1.0"
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": "3.8.0"
				},
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "dummy-driver",
					"version": "1.1.0"
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExcludeLibraryBlankVersion(t *testing.T) {
	cfg := `
	exclude {
		match_type = "strict"
		library {
			name = "mongo-java-driver"
			version = ""
		}
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Exclude)
	require.Equal(t, "strict", string(otelObj.Exclude.MatchType))

	lib := &otelObj.Exclude.Libraries[0]
	require.Equal(t, "mongo-java-driver", lib.Name)
	require.Equal(t, "", *lib.Version)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": "3.8.0"
				},
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": ""
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "dummy-driver",
					"version": "1.1.0"
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "host.type",
					"value": { "stringValue": "n1-standard-1" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": "3.8.0"
				},
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "mongo-java-driver",
					"version": ""
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcB" }
				}]
			},
			"scopeSpans": [{
				"scope": {
					"name": "dummy-driver",
					"version": "1.1.0"
				},
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExcludeServices(t *testing.T) {
	cfg := `
	exclude {
		match_type = "regexp"
		services = ["auth.*", "login.*"]
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Exclude)
	require.Equal(t, "regexp", string(otelObj.Exclude.MatchType))

	svc := &otelObj.Exclude.Services
	require.Equal(t, "auth.*", (*svc)[0])
	require.Equal(t, "login.*", (*svc)[1])

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "auth.basic" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "login.user" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcC" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "auth.basic" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "login.user" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcC" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "test_request",
						"value": { "stringValue": "req_body" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_SelectiveProcessing(t *testing.T) {
	cfg := `
	include {
		match_type = "strict"
		services = ["svcA", "svcB"]
	}
	exclude {
		match_type = "strict"
		attribute {
			key = "redact_trace"
			value = false
		}
	}
	action {
		key = "credit_card"
		action = "delete"
	}
	action {
		key = "duplicate_key"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Include)
	require.Equal(t, "strict", string(otelObj.Include.MatchType))

	svc := &otelObj.Include.Services
	require.Equal(t, "svcA", (*svc)[0])
	require.Equal(t, "svcB", (*svc)[1])

	require.NotNil(t, otelObj.Exclude)
	require.Equal(t, "strict", string(otelObj.Exclude.MatchType))

	attr := &otelObj.Exclude.Attributes[0]
	require.Equal(t, "redact_trace", attr.Key)
	require.Equal(t, false, attr.Value)

	action := &otelObj.Actions[0]
	require.Equal(t, "credit_card", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "duplicate_key", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "redact_trace",
						"value": { "boolValue": true }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "login.user" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "redact_trace",
						"value": { "boolValue": true }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "redact_trace",
						"value": { "boolValue": true }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "login.user" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "redact_trace",
						"value": { "boolValue": true }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Complex(t *testing.T) {
	cfg := `
	action {
		key = "operation"
		value = "default"
		action = "insert"
	}
	action {
		key = "svc.operation"
		value = "operation"
		action = "upsert"
	}
	action {
		key = "operation"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "operation", action.Key)
	require.Equal(t, "default", action.Value)
	require.Equal(t, "insert", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "svc.operation", action.Key)
	require.Equal(t, "operation", action.Value)
	require.Equal(t, "upsert", string(action.Action))

	action = &otelObj.Actions[2]
	require.Equal(t, "operation", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "svc.operation",
						"value": { "stringValue": "old_operation" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "svcA" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "env",
						"value": { "stringValue": "dev" }
					},
					{
						"key": "svc.operation",
						"value": { "stringValue": "operation" }
					},
					{
						"key": "credit_card",
						"value": { "stringValue": "0000-00000-00000" }
					},
					{
						"key": "duplicate_key",
						"value": { "stringValue": "deuplicateduplicatekey" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_ExampleActions(t *testing.T) {
	cfg := `
	action {
		key = "db.table"
		action = "delete"
	}
	action {
		key = "redacted_span"
		value = true
		action = "upsert"
	}
	action {
		key = "copy_key"
		from_attribute = "key_original"
		action = "update"
	}
	action {
		key = "account_id"
		value = 2245
		action = "insert"
	}
	action {
		key = "account_password"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "db.table", action.Key)
	require.Equal(t, "delete", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "redacted_span", action.Key)
	require.Equal(t, true, action.Value)
	require.Equal(t, "upsert", string(action.Action))

	action = &otelObj.Actions[2]
	require.Equal(t, "copy_key", action.Key)
	require.Equal(t, "key_original", action.FromAttribute)
	require.Equal(t, "update", string(action.Action))

	action = &otelObj.Actions[3]
	require.Equal(t, "account_id", action.Key)
	require.Equal(t, 2245, otelObj.Actions[3].Value)
	require.Equal(t, "insert", string(action.Action))

	action = &otelObj.Actions[4]
	require.Equal(t, "account_password", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "db.table",
						"value": { "stringValue": "users" }
					},
					{
						"key": "key_original",
						"value": { "stringValue": "original_data" }
					},
					{
						"key": "copy_key",
						"value": { "stringValue": "non_original_data" }
					},
					{
						"key": "account_password",
						"value": { "stringValue": "12345" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "account_id",
						"value": { "intValue": "2245" }
					},
					{
						"key": "key_original",
						"value": { "stringValue": "original_data" }
					},
					{
						"key": "copy_key",
						"value": { "stringValue": "original_data" }
					},
					{
						"key": "redacted_span",
						"value": { "boolValue": true }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Regexp(t *testing.T) {
	cfg := `
	include {
		match_type = "regexp"
		services = ["auth.*"]
	}
	exclude {
		match_type = "regexp"
		span_names = ["login.*"]
	}
	action {
		key = "password"
		action = "update"
		value = "obfuscated"
	}
	action {
		key = "token"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Include)
	require.Equal(t, "regexp", string(otelObj.Include.MatchType))
	require.Equal(t, "auth.*", otelObj.Include.Services[0])

	require.NotNil(t, otelObj.Exclude)
	require.Equal(t, "regexp", string(otelObj.Exclude.MatchType))
	require.Equal(t, "login.*", otelObj.Exclude.SpanNames[0])

	action := &otelObj.Actions[0]
	require.Equal(t, "password", action.Key)
	require.Equal(t, "obfuscated", action.Value)
	require.Equal(t, "update", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "token", action.Key)
	require.Equal(t, "delete", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "auth.basic" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "12345" }
					},
					{
						"key": "token",
						"value": { "stringValue": "secret_token" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "login.user" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "12345" }
					},
					{
						"key": "token",
						"value": { "stringValue": "secret_token" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "auth.basic" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "obfuscated" }
					}]
				}]
			}]
		},
		{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "login.user" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "12345" }
					},
					{
						"key": "token",
						"value": { "stringValue": "secret_token" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_Regexp2(t *testing.T) {
	cfg := `
	include {
		match_type = "regexp"
		attribute {
			key = "db.statement"
			value = "SELECT \\* FROM USERS.*"
		}
	}
	action {
		key = "db.statement"
		action = "update"
		value = "SELECT * FROM USERS [obfuscated]"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Include)
	require.Equal(t, "regexp", string(otelObj.Include.MatchType))

	attr := &otelObj.Include.Attributes[0]
	require.Equal(t, "db.statement", attr.Key)
	require.Equal(t, "SELECT \\* FROM USERS.*", attr.Value)

	action := &otelObj.Actions[0]
	require.Equal(t, "db.statement", action.Key)
	require.Equal(t, "SELECT * FROM USERS [obfuscated]", action.Value)
	require.Equal(t, "update", string(action.Action))

	var inputTrace = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM USERS_SECRETS" }
					}]
				}]
			}]
		},
		{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM PRODUCTS" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputTrace := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan1",
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM USERS [obfuscated]" }
					}]
				}]
			}]
		},
		{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan2",
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM PRODUCTS" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewTraceSignal(inputTrace, expectedOutputTrace))
}

func Test_LogBodyRegexp(t *testing.T) {
	cfg := `
	include {
		match_type = "regexp"
		log_bodies = ["AUTH.*"]
	}
	action {
		key = "password"
		action = "update"
		value = "obfuscated"
	}
	action {
		key = "token"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Include)
	require.Equal(t, "regexp", string(otelObj.Include.MatchType))

	require.Equal(t, "AUTH.*", otelObj.Include.LogBodies[0])

	action := &otelObj.Actions[0]
	require.Equal(t, "password", action.Key)
	require.Equal(t, "obfuscated", action.Value)
	require.Equal(t, "update", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "token", otelObj.Actions[1].Key)
	require.Equal(t, "delete", string(action.Action))

	var inputLog = `{
		"resourceLogs": [{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000111",
					"severityNumber": 9,
					"severityText": "Info",
					"name": "logA",
					"body": { "stringValue": "AUTH log message" },
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM USERS_SECRETS" }
					}]
				}]
			}]
		},
		{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000999",
					"severityNumber": 9,
					"severityText": "Info",
					"name": "logA",
					"body": { "stringValue": "This is a log message" },
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM USERS_SECRETS" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputLog := `{
		"resourceLogs": [{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000111",
					"severityNumber": 9,
					"severityText": "Info",
					"name": "logA",
					"body": { "stringValue": "AUTH log message" },
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM USERS_SECRETS" }
					}]
				}]
			}]
		},
		{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000999",
					"severityNumber": 9,
					"severityText": "Info",
					"name": "logA",
					"body": { "stringValue": "This is a log message" },
					"attributes": [{
						"key": "db.statement",
						"value": { "stringValue": "SELECT * FROM USERS_SECRETS" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewLogSignal(inputLog, expectedOutputLog))
}

func Test_LogSeverityRegexp(t *testing.T) {
	cfg := `
	include {
		match_type = "regexp"
		log_severity_texts = ["info.*"]
		log_severity {
			min = "INFO"
			match_undefined = true
		}
	}
	action {
		key = "password"
		action = "update"
		value = "obfuscated"
	}
	action {
		key = "token"
		action = "delete"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Include)
	require.Equal(t, "regexp", string(otelObj.Include.MatchType))

	require.Equal(t, "info.*", otelObj.Include.LogSeverityTexts[0])

	require.Equal(t, int32(9), int32(otelObj.Include.LogSeverityNumber.Min))
	require.Equal(t, true, otelObj.Include.LogSeverityNumber.MatchUndefined)

	action := &otelObj.Actions[0]
	require.Equal(t, "password", action.Key)
	require.Equal(t, "obfuscated", action.Value)
	require.Equal(t, "update", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "token", otelObj.Actions[1].Key)
	require.Equal(t, "delete", string(action.Action))

	var inputLog = `{
		"resourceLogs": [{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000111",
					"severityNumber": 9,
					"severityText": "info",
					"name": "logA",
					"body": { "stringValue": "AUTH log message" },
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "12345" }
					},
					{
						"key": "token",
						"value": { "stringValue": "12345" }
					}]
				}]
			}]
		},
		{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000999",
					"severityNumber": 5,
					"severityText": "debug",
					"name": "logA",
					"body": { "stringValue": "AUTH log message" },
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "12345" }
					},
					{
						"key": "token",
						"value": { "stringValue": "12345" }
					}]
				}]
			}]
		}]
	}`

	expectedOutputLog := `{
		"resourceLogs": [{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000111",
					"severityNumber": 9,
					"severityText": "info",
					"name": "logA",
					"body": { "stringValue": "AUTH log message" },
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "obfuscated" }
					}]
				}]
			}]
		},
		{
			"scopeLogs": [{
				"log_records": [{
					"timeUnixNano": "1581452773000000999",
					"severityNumber": 5,
					"severityText": "debug",
					"name": "logA",
					"body": { "stringValue": "AUTH log message" },
					"attributes": [{
						"key": "password",
						"value": { "stringValue": "12345" }
					},
					{
						"key": "token",
						"value": { "stringValue": "12345" }
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewLogSignal(inputLog, expectedOutputLog))
}

func Test_FromContext(t *testing.T) {
	cfg := `
	action {
		key = "origin"
		from_context = "metadata.origin"
		action = "insert"
	}
	action {
		key = "enduser.id"
		from_context = "auth.subject"
		action = "insert"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)

	otelObj := (convertedArgs).(*attributesprocessor.Config)

	action := &otelObj.Actions[0]
	require.Equal(t, "origin", action.Key)
	require.Equal(t, "metadata.origin", action.FromContext)
	require.Equal(t, "insert", string(action.Action))

	action = &otelObj.Actions[1]
	require.Equal(t, "enduser.id", action.Key)
	require.Equal(t, "auth.subject", action.FromContext)
	require.Equal(t, "insert", string(action.Action))
}

func Test_MetricNames(t *testing.T) {
	cfg := `
	include {
		match_type = "regexp"
		metric_names = ["counter.*"]
	}
	action {
		key = "important_label"
		action = "upsert"
		value = "label_val"
	}

	output {
		// no-op: will be overridden by test code.
	}
	`
	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*attributesprocessor.Config)

	require.NotNil(t, otelObj.Include)
	require.Equal(t, "regexp", string(otelObj.Include.MatchType))

	require.Equal(t, "counter.*", otelObj.Include.MetricNames[0])

	action := &otelObj.Actions[0]
	require.Equal(t, "important_label", action.Key)
	require.Equal(t, "label_val", action.Value)
	require.Equal(t, "upsert", string(action.Action))

	var inputMetric = `{
		"resourceMetrics": [{
			"scopeMetrics": [{
				"metrics": [{
					"name": "counter-int",
					"unit": "1",
					"sum": {
						"dataPoints": [{
							"attributes": [{
								"key": "label-1",
								"value": { "stringValue": "label-value-1" }
							},
							{
								"key": "label2",
								"value": { "stringValue": "label-value-2" }
							}],
							"startTimeUnixNano": "1581452773000000789",
							"timeUnixNano": "1581452773000000789",
							"asInt": "123"
						}],
						"aggregationTemporality": 2,
						"isMonotonic": true
					}
				}]
			}]
		},
		{
			"scopeMetrics": [{
				"metrics": [{
					"name": "c-int",
					"unit": "1",
					"sum": {
						"dataPoints": [{
							"attributes": [{
								"key": "label-1",
								"value": { "stringValue": "label-value-1" }
							},
							{
								"key": "label2",
								"value": { "stringValue": "label-value-2" }
							}],
							"startTimeUnixNano": "1581452772000000321",
							"timeUnixNano": "1581452773000000789",
							"asInt": "456"
						}],
						"aggregationTemporality": 2,
						"isMonotonic": true
					}
				}]
			}]
		}]
	}`

	expectedOutputMetric := `{
		"resourceMetrics": [{
			"scopeMetrics": [{
				"metrics": [{
					"name": "counter-int",
					"unit": "1",
					"sum": {
						"dataPoints": [{
							"attributes": [{
								"key": "label-1",
								"value": { "stringValue": "label-value-1" }
							},
							{
								"key": "label2",
								"value": { "stringValue": "label-value-2" }
							},
							{
								"key": "important_label",
								"value": { "stringValue": "label_val" }
							}],
							"startTimeUnixNano": "1581452773000000789",
							"timeUnixNano": "1581452773000000789",
							"asInt": "123"
						}],
						"aggregationTemporality": 2,
						"isMonotonic": true
					}
				}]
			}]
		},
		{
			"scopeMetrics": [{
				"metrics": [{
					"name": "c-int",
					"unit": "1",
					"sum": {
						"dataPoints": [{
							"attributes": [{
								"key": "label-1",
								"value": { "stringValue": "label-value-1" }
							},
							{
								"key": "label2",
								"value": { "stringValue": "label-value-2" }
							}],
							"startTimeUnixNano": "1581452772000000321",
							"timeUnixNano": "1581452773000000789",
							"asInt": "456"
						}],
						"aggregationTemporality": 2,
						"isMonotonic": true
					}
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, NewMetricSignal(inputMetric, expectedOutputMetric))
}

type signal interface {
	MakeOutput() *otelcol.ConsumerArguments
	ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error
	CheckOutput(t *testing.T)
}

type traceSignal struct {
	traceCh              chan ptrace.Traces
	inputTrace           ptrace.Traces
	expectedOuutputTrace ptrace.Traces
}

func NewTraceSignal(inputJson string, expectedOutputJson string) signal {
	return &traceSignal{
		traceCh:              make(chan ptrace.Traces),
		inputTrace:           createTestTraces(inputJson),
		expectedOuutputTrace: createTestTraces(expectedOutputJson),
	}
}

func (s traceSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeTracesOutput(s.traceCh)
}

func (s traceSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeTraces(ctx, s.inputTrace)
}

func (s traceSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to traceCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for traces")
	case tr := <-s.traceCh:
		trStr := marshalTraces(tr)
		expStr := marshalTraces(s.expectedOuutputTrace)
		require.JSONEq(t, expStr, trStr)
	}
}

type logSignal struct {
	logCh              chan plog.Logs
	inputLog           plog.Logs
	expectedOuutputLog plog.Logs
}

func NewLogSignal(inputJson string, expectedOutputJson string) signal {
	return &logSignal{
		logCh:              make(chan plog.Logs),
		inputLog:           createTestLogs(inputJson),
		expectedOuutputLog: createTestLogs(expectedOutputJson),
	}
}

func (s logSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeLogsOutput(s.logCh)
}

func (s logSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeLogs(ctx, s.inputLog)
}

func (s logSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to logCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for logs")
	case tr := <-s.logCh:
		trStr := marshalLogs(tr)
		expStr := marshalLogs(s.expectedOuutputLog)
		require.JSONEq(t, expStr, trStr)
	}
}

type metricSignal struct {
	metricCh              chan pmetric.Metrics
	inputMetric           pmetric.Metrics
	expectedOuutputMetric pmetric.Metrics
}

func NewMetricSignal(inputJson string, expectedOutputJson string) signal {
	return &metricSignal{
		metricCh:              make(chan pmetric.Metrics),
		inputMetric:           createTestMetrics(inputJson),
		expectedOuutputMetric: createTestMetrics(expectedOutputJson),
	}
}

func (s metricSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeMetricsOutput(s.metricCh)
}

func (s metricSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeMetrics(ctx, s.inputMetric)
}

func (s metricSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to logCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for logs")
	case tr := <-s.metricCh:
		trStr := marshalMetrics(tr)
		expStr := marshalMetrics(s.expectedOuutputMetric)
		require.JSONEq(t, expStr, trStr)
	}
}

func testRunProcessor(t *testing.T, processorConfig string, testSignal signal) {
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.attributes")
	require.NoError(t, err)

	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(processorConfig), &args))

	// Override the arguments so signals get forwarded to the test channel.
	args.Output = testSignal.MakeOutput()

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

	// Send signals in the background to our processor.
	go func() {
		exports := ctrl.Exports().(otelcol.ConsumerExports)

		bo := backoff.New(ctx, backoff.Config{
			MinBackoff: 10 * time.Millisecond,
			MaxBackoff: 100 * time.Millisecond,
		})
		for bo.Ongoing() {
			err := testSignal.ConsumeInput(ctx, exports.Input)
			if err != nil {
				level.Error(l).Log("msg", "failed to send traces", "err", err)
				bo.Wait()
				continue
			}

			return
		}
	}()

	testSignal.CheckOutput(t)
}

// makeTracesOutput returns ConsumerArguments which will forward traces to the
// provided channel.
func makeTracesOutput(ch chan ptrace.Traces) *otelcol.ConsumerArguments {
	traceConsumer := fakeconsumer.Consumer{
		ConsumeTracesFunc: func(ctx context.Context, t ptrace.Traces) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Traces: []otelcol.Consumer{&traceConsumer},
	}
}

// traceJson should match format from the protobuf definition:
// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto
func createTestTraces(traceJson string) ptrace.Traces {
	decoder := &ptrace.JSONUnmarshaler{}
	data, err := decoder.UnmarshalTraces([]byte(traceJson))
	if err != nil {
		panic(err)
	}
	return data
}

func marshalTraces(trace ptrace.Traces) string {
	marshaler := &ptrace.JSONMarshaler{}
	data, err := marshaler.MarshalTraces(trace)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// makeLogsOutput returns ConsumerArguments which will forward logs to the
// provided channel.
func makeLogsOutput(ch chan plog.Logs) *otelcol.ConsumerArguments {
	logConsumer := fakeconsumer.Consumer{
		ConsumeLogsFunc: func(ctx context.Context, t plog.Logs) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Logs: []otelcol.Consumer{&logConsumer},
	}
}

// logJson should match format from the protobuf definition:
// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
func createTestLogs(logJson string) plog.Logs {
	decoder := &plog.JSONUnmarshaler{}
	data, err := decoder.UnmarshalLogs([]byte(logJson))
	if err != nil {
		panic(err)
	}
	return data
}

func marshalLogs(log plog.Logs) string {
	marshaler := &plog.JSONMarshaler{}
	data, err := marshaler.MarshalLogs(log)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// makeMetricsOutput returns ConsumerArguments which will forward metrics to the
// provided channel.
func makeMetricsOutput(ch chan pmetric.Metrics) *otelcol.ConsumerArguments {
	metricConsumer := fakeconsumer.Consumer{
		ConsumeMetricsFunc: func(ctx context.Context, t pmetric.Metrics) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Metrics: []otelcol.Consumer{&metricConsumer},
	}
}

// metricJson should match format from the protobuf definition:
// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/metrics/v1/metrics.proto
func createTestMetrics(metricJson string) pmetric.Metrics {
	decoder := &pmetric.JSONUnmarshaler{}
	data, err := decoder.UnmarshalMetrics([]byte(metricJson))
	if err != nil {
		panic(err)
	}
	return data
}

func marshalMetrics(metrics pmetric.Metrics) string {
	marshaler := &pmetric.JSONMarshaler{}
	data, err := marshaler.MarshalMetrics(metrics)
	if err != nil {
		panic(err)
	}
	return string(data)
}
