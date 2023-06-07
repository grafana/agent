package convert_test

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/otelcol/exporter/loki/internal/convert"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestConverter(t *testing.T) {
	tt := []struct {
		name            string
		input           string
		expectLine      string
		expectLabels    string
		expectTimestamp time.Time
	}{
		{
			name: "log line without format hint",
			input: `{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "host.name",
            "value": {
              "stringValue": "testHost"
            }
          }
        ],
        "droppedAttributesCount": 1
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "name",
            "version": "version",
            "droppedAttributesCount": 1
          },
          "logRecords": [
            {
              "timeUnixNano": "1672827031972869000",
              "observedTimeUnixNano": "1672827031972869000",
              "severityNumber": 17,
              "severityText": "Error",
              "body": {
                "stringValue": "hello world"
              },
              "attributes": [
                {
                  "key": "sdkVersion",
                  "value": {
                    "stringValue": "1.0.1"
                  }
                }
              ],
              "droppedAttributesCount": 1,
              "traceId": "0102030405060708090a0b0c0d0e0f10",
              "spanId": "1112131415161718"
            }
          ],
          "schemaUrl": "ScopeLogsSchemaURL"
        }
      ],
      "schemaUrl": "testSchemaURL"
    }
  ]
}`,
			expectLine:      `{"body":"hello world","traceid":"0102030405060708090a0b0c0d0e0f10","spanid":"1112131415161718","severity":"Error","attributes":{"sdkVersion":"1.0.1"},"resources":{"host.name":"testHost"},"instrumentation_scope":{"name":"name","version":"version"}}`,
			expectLabels:    `{exporter="OTLP", level="ERROR"}`,
			expectTimestamp: time.Date(2023, time.January, 4, 10, 10, 31, 972869000, time.UTC),
		},
		{
			name: "log line with logfmt format hint in resource attributes",
			input: `{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "host.name",
            "value": {
              "stringValue": "testHost"
            }
          },
          {
            "key": "loki.format",
            "value": {
              "stringValue": "logfmt"
            }
          }
        ],
        "droppedAttributesCount": 1
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "name",
            "version": "version",
            "droppedAttributesCount": 1
          },
          "logRecords": [
            {
              "timeUnixNano": "1672827031972869000",
              "observedTimeUnixNano": "1672827031972869000",
              "severityNumber": 17,
              "severityText": "Error",
              "body": {
                "stringValue": "msg=\"hello world\""
              },
              "attributes": [
                {
                  "key": "sdkVersion",
                  "value": {
                    "stringValue": "1.0.1"
                  }
                }
              ],
              "droppedAttributesCount": 1,
              "traceId": "0102030405060708090a0b0c0d0e0f10",
              "spanId": "1112131415161718"
            }
          ],
          "schemaUrl": "ScopeLogsSchemaURL"
        }
      ],
      "schemaUrl": "testSchemaURL"
    }
  ]
}`,
			expectLine:      `msg="hello world" traceID=0102030405060708090a0b0c0d0e0f10 spanID=1112131415161718 severity=Error attribute_sdkVersion=1.0.1 resource_host.name=testHost instrumentation_scope_name=name instrumentation_scope_version=version`,
			expectLabels:    `{exporter="OTLP", level="ERROR"}`,
			expectTimestamp: time.Date(2023, time.January, 4, 10, 10, 31, 972869000, time.UTC),
		},
		{
			name: "log line with logfmt format hint in log attributes",
			input: `{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "host.name",
            "value": {
              "stringValue": "testHost"
            }
          }
        ],
        "droppedAttributesCount": 1
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "name",
            "version": "version",
            "droppedAttributesCount": 1
          },
          "logRecords": [
            {
              "timeUnixNano": "1672827031972869000",
              "observedTimeUnixNano": "1672827031972869000",
              "severityNumber": 17,
              "severityText": "Error",
              "body": {
                "stringValue": "msg=\"hello world\""
              },
              "attributes": [
                {
                  "key": "sdkVersion",
                  "value": {
                    "stringValue": "1.0.1"
                  }
                },
                {
                  "key": "loki.format",
                  "value": {
                    "stringValue": "logfmt"
                  }
                }
              ],
              "droppedAttributesCount": 1,
              "traceId": "0102030405060708090a0b0c0d0e0f10",
              "spanId": "1112131415161718"
            }
          ],
          "schemaUrl": "ScopeLogsSchemaURL"
        }
      ],
      "schemaUrl": "testSchemaURL"
    }
  ]
}`,
			expectLine:      `msg="hello world" traceID=0102030405060708090a0b0c0d0e0f10 spanID=1112131415161718 severity=Error attribute_sdkVersion=1.0.1 resource_host.name=testHost instrumentation_scope_name=name instrumentation_scope_version=version`,
			expectLabels:    `{exporter="OTLP", level="ERROR"}`,
			expectTimestamp: time.Date(2023, time.January, 4, 10, 10, 31, 972869000, time.UTC),
		},
		{
			name: "resource attributes converted to labels",
			input: `{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "host.name",
            "value": {
              "stringValue": "testHost"
            }
          },
		  {
		    "key": "loki.resource.labels",
			"value": {
			  "stringValue": "mylabel_1,mylabel_2"
			}
		  },
		  {
		    "key": "mylabel_1",
			"value": {
			  "stringValue": "value_1"
			}
		  },
		  {
		    "key": "mylabel_2",
			"value": {
			  "intValue": "42"
			}
		  },
		  {
		    "key": "mylabel_3",
			"value": {
			  "stringValue": "value_3"
			}
		  }
        ],
        "droppedAttributesCount": 1
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "name",
            "version": "version",
            "droppedAttributesCount": 1
          },
          "logRecords": [
            {
              "timeUnixNano": "1672827031972869000",
              "observedTimeUnixNano": "1672827031972869000",
              "severityNumber": 17,
              "severityText": "Error",
              "body": {
                "stringValue": "msg=\"hello world\""
              },
              "attributes": [
                {
                  "key": "sdkVersion",
                  "value": {
                    "stringValue": "1.0.1"
                  }
                },
                {
                  "key": "loki.format",
                  "value": {
                    "stringValue": "logfmt"
                  }
                }
              ],
              "droppedAttributesCount": 1,
              "traceId": "0102030405060708090a0b0c0d0e0f10",
              "spanId": "1112131415161718"
            }
          ],
          "schemaUrl": "ScopeLogsSchemaURL"
        }
      ],
      "schemaUrl": "testSchemaURL"
    }
  ]
}`,
			expectLine:      `msg="hello world" traceID=0102030405060708090a0b0c0d0e0f10 spanID=1112131415161718 severity=Error attribute_sdkVersion=1.0.1 resource_host.name=testHost resource_mylabel_3=value_3 instrumentation_scope_name=name instrumentation_scope_version=version`,
			expectLabels:    `{exporter="OTLP", level="ERROR", mylabel_1="value_1", mylabel_2="42"}`,
			expectTimestamp: time.Date(2023, time.January, 4, 10, 10, 31, 972869000, time.UTC),
		},
		{
			name: "log attributes converted to labels",
			input: `{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "host.name",
            "value": {
              "stringValue": "testHost"
            }
          }
        ],
        "droppedAttributesCount": 1
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "name",
            "version": "version",
            "droppedAttributesCount": 1
          },
          "logRecords": [
            {
              "timeUnixNano": "1672827031972869000",
              "observedTimeUnixNano": "1672827031972869000",
              "severityNumber": 17,
              "severityText": "Error",
              "body": {
                "stringValue": "msg=\"hello world\""
              },
              "attributes": [
                {
                  "key": "sdkVersion",
                  "value": {
                    "stringValue": "1.0.1"
                  }
                },
                {
                  "key": "loki.format",
                  "value": {
                    "stringValue": "logfmt"
                  }
                },
		        {
		          "key": "loki.attribute.labels",
		          "value": {
		            "stringValue": "mylabel_1,mylabel_2"
		          }
		        },
		        {
		          "key": "mylabel_1",
		          "value": {
		            "stringValue": "value_1"
		          }
		        },
		        {
		          "key": "mylabel_2",
		          "value": {
		            "intValue": "42"
		          }
		        },
		        {
		          "key": "mylabel_3",
		          "value": {
		            "stringValue": "value_3"
		          }
		        }
              ],
              "droppedAttributesCount": 1,
              "traceId": "0102030405060708090a0b0c0d0e0f10",
              "spanId": "1112131415161718"
            }
          ],
          "schemaUrl": "ScopeLogsSchemaURL"
        }
      ],
      "schemaUrl": "testSchemaURL"
    }
  ]
}`,
			expectLine:      `msg="hello world" traceID=0102030405060708090a0b0c0d0e0f10 spanID=1112131415161718 severity=Error attribute_sdkVersion=1.0.1 attribute_mylabel_3=value_3 resource_host.name=testHost instrumentation_scope_name=name instrumentation_scope_version=version`,
			expectLabels:    `{exporter="OTLP", level="ERROR", mylabel_1="value_1", mylabel_2="42"}`,
			expectTimestamp: time.Date(2023, time.January, 4, 10, 10, 31, 972869000, time.UTC),
		},
		{
			name: "tenant resource attribute converted to label",
			input: `{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "host.name",
            "value": {
              "stringValue": "testHost"
            }
          }
        ],
        "droppedAttributesCount": 1
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "name",
            "version": "version",
            "droppedAttributesCount": 1
          },
          "logRecords": [
            {
              "timeUnixNano": "1672827031972869000",
              "observedTimeUnixNano": "1672827031972869000",
              "severityNumber": 17,
              "severityText": "Error",
              "body": {
                "stringValue": "hello world"
              },
              "attributes": [
                {
                  "key": "sdkVersion",
                  "value": {
                    "stringValue": "1.0.1"
                  }
                },
                {
                  "key": "loki.format",
                  "value": {
                    "stringValue": "json"
                  }
                },
                {
                  "key": "loki.tenant",
                  "value": {
                    "stringValue": "tenant.id"
                  }
                },
                {
                  "key": "tenant.id",
                  "value": {
                    "stringValue": "tenant_2"
                  }
                }
              ],
              "droppedAttributesCount": 1,
              "traceId": "0102030405060708090a0b0c0d0e0f10",
              "spanId": "1112131415161718"
            }
          ],
          "schemaUrl": "ScopeLogsSchemaURL"
        }
      ],
      "schemaUrl": "testSchemaURL"
    }
  ]
}`,
			expectLine:      `{"body":"hello world","traceid":"0102030405060708090a0b0c0d0e0f10","spanid":"1112131415161718","severity":"Error","attributes":{"sdkVersion":"1.0.1"},"resources":{"host.name":"testHost"},"instrumentation_scope":{"name":"name","version":"version"}}`,
			expectLabels:    `{exporter="OTLP", level="ERROR", tenant.id="tenant_2"}`,
			expectTimestamp: time.Date(2023, time.January, 4, 10, 10, 31, 972869000, time.UTC),
		},
	}

	decoder := &plog.JSONUnmarshaler{}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := decoder.UnmarshalLogs([]byte(tc.input))
			require.NoError(t, err)

			l := util.TestLogger(t)
			ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)
			conv := convert.New(l, prometheus.NewRegistry(), []loki.LogsReceiver{ch1, ch2})
			go func() {
				require.NoError(t, conv.ConsumeLogs(context.Background(), payload))
			}()

			for i := 0; i < 2; i++ {
				select {
				case l := <-ch1:
					require.Equal(t, tc.expectLine, l.Line)
					require.Equal(t, tc.expectLabels, l.Labels.String())
					require.Equal(t, tc.expectTimestamp, l.Timestamp.UTC())
				case l := <-ch2:
					require.Equal(t, tc.expectLine, l.Line)
					require.Equal(t, tc.expectLabels, l.Labels.String())
					require.Equal(t, tc.expectTimestamp, l.Timestamp.UTC())
				case <-time.After(time.Second):
					require.FailNow(t, "failed waiting for logs")
				}
			}
		})
	}
}
