package transform_test

import (
	"testing"

	"github.com/grafana/agent/component/otelcol/processor/transform"
	"github.com/grafana/river"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		expected map[string]interface{}
		errorMsg string
	}{
		{
			testName: "Defaults",
			cfg: `
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "propagate",
			},
		},
		{
			testName: "IgnoreErrors",
			cfg: `
			error_mode = "ignore"
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
			},
		},
		{
			testName: "TransformIfFieldDoesNotExist",
			cfg: `
			error_mode = "ignore"
			trace_statements {
				context = "span"
				statements = [
					// Accessing a map with a key that does not exist will return nil. 
					"set(attributes[\"test\"], \"pass\") where attributes[\"test\"] == nil",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"trace_statements": []interface{}{
					map[string]interface{}{
						"context": "span",
						"statements": []interface{}{
							`set(attributes["test"], "pass") where attributes["test"] == nil`,
						},
					},
				},
			},
		},
		{
			testName: "RenameAttribute1",
			cfg: `
			error_mode = "ignore"
			trace_statements {
				context = "resource"
				statements = [
					"set(attributes[\"namespace\"], attributes[\"k8s.namespace.name\"])",
					"delete_key(attributes, \"k8s.namespace.name\")",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"trace_statements": []interface{}{
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`set(attributes["namespace"], attributes["k8s.namespace.name"])`,
							`delete_key(attributes, "k8s.namespace.name")`,
						},
					},
				},
			},
		},
		{
			testName: "RenameAttribute2",
			cfg: `
			error_mode = "ignore"
			trace_statements {
				context = "resource"
				statements = [
					"replace_all_patterns(attributes, \"key\", \"k8s\\\\.namespace\\\\.name\", \"namespace\")",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"trace_statements": []interface{}{
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`replace_all_patterns(attributes, "key", "k8s\\.namespace\\.name", "namespace")`,
						},
					},
				},
			},
		},
		{
			testName: "CreateAttributeFromContentOfLogBody",
			cfg: `
			error_mode = "ignore"
			log_statements {
				context = "log"
				statements = [
					"set(attributes[\"body\"], body)",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"log_statements": []interface{}{
					map[string]interface{}{
						"context": "log",
						"statements": []interface{}{
							`set(attributes["body"], body)`,
						},
					},
				},
			},
		},
		{
			testName: "CombineTwoAttributes",
			cfg: `
			error_mode = "ignore"
			trace_statements {
				context = "resource"
				statements = [
					// The Concat function combines any number of strings, separated by a delimiter.
					"set(attributes[\"test\"], Concat([attributes[\"foo\"], attributes[\"bar\"]], \" \"))",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"trace_statements": []interface{}{
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`set(attributes["test"], Concat([attributes["foo"], attributes["bar"]], " "))`,
						},
					},
				},
			},
		},
		{
			testName: "ParseJsonLogs",
			cfg: `
			error_mode = "ignore"
			log_statements {
				context = "log"
				statements = [
					"merge_maps(cache, ParseJSON(body), \"upsert\") where IsMatch(body, \"^\\\\{\") ",
					"set(attributes[\"attr1\"], cache[\"attr1\"])",
					"set(attributes[\"attr2\"], cache[\"attr2\"])",
					"set(attributes[\"nested.attr3\"], cache[\"nested\"][\"attr3\"])",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"log_statements": []interface{}{
					map[string]interface{}{
						"context": "log",
						"statements": []interface{}{
							`merge_maps(cache, ParseJSON(body), "upsert") where IsMatch(body, "^\\{") `,
							`set(attributes["attr1"], cache["attr1"])`,
							`set(attributes["attr2"], cache["attr2"])`,
							`set(attributes["nested.attr3"], cache["nested"]["attr3"])`,
						},
					},
				},
			},
		},
		{
			testName: "ManyStatements1",
			cfg: `
			error_mode = "ignore"
			trace_statements {
				context = "resource"
				statements = [
					"keep_keys(attributes, [\"service.name\", \"service.namespace\", \"cloud.region\", \"process.command_line\"])",
					"replace_pattern(attributes[\"process.command_line\"], \"password\\\\=[^\\\\s]*(\\\\s?)\", \"password=***\")",
					"limit(attributes, 100, [])",
					"truncate_all(attributes, 4096)",
				]
			}
			trace_statements {
				context = "span"
				statements = [
					"set(status.code, 1) where attributes[\"http.path\"] == \"/health\"",
					"set(name, attributes[\"http.route\"])",
					"replace_match(attributes[\"http.target\"], \"/user/*/list/*\", \"/user/{userId}/list/{listId}\")",
					"limit(attributes, 100, [])",
					"truncate_all(attributes, 4096)",
				]
			}
			metric_statements {
				context = "resource"
				statements = [
					"keep_keys(attributes, [\"host.name\"])",
					"truncate_all(attributes, 4096)",
				]
			}
			metric_statements {
				context = "metric"
				statements = [
					"set(description, \"Sum\") where type == \"Sum\"",
				]
			}
			metric_statements {
				context = "datapoint"
				statements = [
					"limit(attributes, 100, [\"host.name\"])",
					"truncate_all(attributes, 4096)",
					"convert_sum_to_gauge() where metric.name == \"system.processes.count\"",
					"convert_gauge_to_sum(\"cumulative\", false) where metric.name == \"prometheus_metric\"",
				]
			}
			log_statements {
				context = "resource"
				statements = [
					"keep_keys(attributes, [\"service.name\", \"service.namespace\", \"cloud.region\"])",
				]
			}
			log_statements {
				context = "log"
				statements = [
					"set(severity_text, \"FAIL\") where body == \"request failed\"",
					"replace_all_matches(attributes, \"/user/*/list/*\", \"/user/{userId}/list/{listId}\")",
					"replace_all_patterns(attributes, \"value\", \"/account/\\\\d{4}\", \"/account/{accountId}\")",
					"set(body, attributes[\"http.route\"])",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "ignore",
				"trace_statements": []interface{}{
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`keep_keys(attributes, ["service.name", "service.namespace", "cloud.region", "process.command_line"])`,
							`replace_pattern(attributes["process.command_line"], "password\\=[^\\s]*(\\s?)", "password=***")`,
							`limit(attributes, 100, [])`,
							`truncate_all(attributes, 4096)`,
						},
					},
					map[string]interface{}{
						"context": "span",
						"statements": []interface{}{
							`set(status.code, 1) where attributes["http.path"] == "/health"`,
							`set(name, attributes["http.route"])`,
							`replace_match(attributes["http.target"], "/user/*/list/*", "/user/{userId}/list/{listId}")`,
							`limit(attributes, 100, [])`,
							`truncate_all(attributes, 4096)`,
						},
					},
				},
				"metric_statements": []interface{}{
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`keep_keys(attributes, ["host.name"])`,
							`truncate_all(attributes, 4096)`,
						},
					},
					map[string]interface{}{
						"context": "metric",
						"statements": []interface{}{
							`set(description, "Sum") where type == "Sum"`,
						},
					},
					map[string]interface{}{
						"context": "datapoint",
						"statements": []interface{}{
							`limit(attributes, 100, ["host.name"])`,
							`truncate_all(attributes, 4096)`,
							`convert_sum_to_gauge() where metric.name == "system.processes.count"`,
							`convert_gauge_to_sum("cumulative", false) where metric.name == "prometheus_metric"`,
						},
					},
				},
				"log_statements": []interface{}{
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`keep_keys(attributes, ["service.name", "service.namespace", "cloud.region"])`,
						},
					},
					map[string]interface{}{
						"context": "log",
						"statements": []interface{}{
							`set(severity_text, "FAIL") where body == "request failed"`,
							`replace_all_matches(attributes, "/user/*/list/*", "/user/{userId}/list/{listId}")`,
							`replace_all_patterns(attributes, "value", "/account/\\d{4}", "/account/{accountId}")`,
							`set(body, attributes["http.route"])`,
						},
					},
				},
			},
		},
		{
			testName: "ManyStatements2",
			cfg: `
			trace_statements {
				context = "span"
				statements = [
					"set(name, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
					"keep_keys(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			trace_statements {
				context = "resource"
				statements = [
					"set(attributes[\"name\"], \"bear\")",
				]
			}
			metric_statements {
				context = "datapoint"
				statements = [
					"set(metric.name, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
					"keep_keys(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			metric_statements {
				context = "resource"
				statements = [
					"set(attributes[\"name\"], \"bear\")",
				]
			}
			log_statements {
				context = "log"
				statements = [
					"set(body, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
					"keep_keys(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			log_statements {
				context = "resource"
				statements = [
					"set(attributes[\"name\"], \"bear\")",
				]
			}
			output {}
			`,
			expected: map[string]interface{}{
				"error_mode": "propagate",
				"trace_statements": []interface{}{
					map[string]interface{}{
						"context": "span",
						"statements": []interface{}{
							`set(name, "bear") where attributes["http.path"] == "/animal"`,
							`keep_keys(attributes, ["http.method", "http.path"])`,
						},
					},
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`set(attributes["name"], "bear")`,
						},
					},
				},
				"metric_statements": []interface{}{
					map[string]interface{}{
						"context": "datapoint",
						"statements": []interface{}{
							`set(metric.name, "bear") where attributes["http.path"] == "/animal"`,
							`keep_keys(attributes, ["http.method", "http.path"])`,
						},
					},
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`set(attributes["name"], "bear")`,
						},
					},
				},
				"log_statements": []interface{}{
					map[string]interface{}{
						"context": "log",
						"statements": []interface{}{
							`set(body, "bear") where attributes["http.path"] == "/animal"`,
							`keep_keys(attributes, ["http.method", "http.path"])`,
						},
					},
					map[string]interface{}{
						"context": "resource",
						"statements": []interface{}{
							`set(attributes["name"], "bear")`,
						},
					},
				},
			},
		},
		{
			testName: "unknown_error_mode",
			cfg: `
			error_mode = "test"
			output {}
			`,
			errorMsg: `2:17: "test" unknown error mode test`,
		},
		{
			testName: "bad_syntax_log",
			cfg: `
			log_statements {
				context = "log"
				statements = [
					"set(body, \"bear\" where attributes[\"http.path\"] == \"/animal\"",
					"keep_keys(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			output {}
			`,
			errorMsg: `unable to parse OTTL statement "set(body, \"bear\" where attributes[\"http.path\"] == \"/animal\"": statement has invalid syntax: 1:18: unexpected token "where" (expected ")" Key*)`,
		},
		{
			testName: "bad_syntax_metric",
			cfg: `
			metric_statements {
				context = "datapoint"
				statements = [
					"set(name, \"bear\" where attributes[\"http.path\"] == \"/animal\"",
					"keep_keys(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			output {}
			`,
			errorMsg: `unable to parse OTTL statement "set(name, \"bear\" where attributes[\"http.path\"] == \"/animal\"": statement has invalid syntax: 1:18: unexpected token "where" (expected ")" Key*)`,
		},
		{
			testName: "bad_syntax_trace",
			cfg: `
			trace_statements {
				context = "span"
				statements = [
					"set(name, \"bear\" where attributes[\"http.path\"] == \"/animal\"",
					"keep_keys(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			output {}
			`,
			errorMsg: `unable to parse OTTL statement "set(name, \"bear\" where attributes[\"http.path\"] == \"/animal\"": statement has invalid syntax: 1:18: unexpected token "where" (expected ")" Key*)`,
		},
		{
			testName: "unknown_function_log",
			cfg: `
			log_statements {
				context = "log"
				statements = [
					"set(body, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
					"not_a_function(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			output {}
			`,
			errorMsg: `unable to parse OTTL statement "not_a_function(attributes, [\"http.method\", \"http.path\"])": undefined function "not_a_function"`,
		},
		{
			testName: "unknown_function_metric",
			cfg: `
			metric_statements {
				context = "datapoint"
				statements = [
					"set(metric.name, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
					"not_a_function(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			output {}
			`,
			errorMsg: `unable to parse OTTL statement "not_a_function(attributes, [\"http.method\", \"http.path\"])": undefined function "not_a_function"`,
		},
		{
			testName: "unknown_function_trace",
			cfg: `
			trace_statements {
				context = "span"
				statements = [
					"set(name, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
					"not_a_function(attributes, [\"http.method\", \"http.path\"])",
				]
			}
			output {}
			`,
			errorMsg: `unable to parse OTTL statement "not_a_function(attributes, [\"http.method\", \"http.path\"])": undefined function "not_a_function"`,
		},
		{
			testName: "unknown_context",
			cfg: `
			trace_statements {
				context = "test"
				statements = [
					"set(name, \"bear\") where attributes[\"http.path\"] == \"/animal\"",
				]
			}
			output {}
			`,
			errorMsg: `3:15: "test" unknown context test`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args transform.Arguments
			err := river.Unmarshal([]byte(tc.cfg), &args)
			if tc.errorMsg != "" {
				require.ErrorContains(t, err, tc.errorMsg)
				return
			}

			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*transformprocessor.Config)

			var expectedCfg transformprocessor.Config
			err = mapstructure.Decode(tc.expected, &expectedCfg)
			require.NoError(t, err)

			// Validate the two configs
			require.NoError(t, actual.Validate())
			require.NoError(t, expectedCfg.Validate())

			// Compare the two configs
			require.Equal(t, expectedCfg, *actual)
		})
	}
}
