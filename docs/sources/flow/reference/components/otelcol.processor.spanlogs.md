---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.spanlogs/
title: otelcol.processor.spanlogs
---

# otelcol.processor.spanlogs

`otelcol.processor.spanlogs` accepts traces telemetry data from other `otelcol`
components and outputs logs telemetry data for each span, root, or process.
This allows for automatically building a mechanism for trace
discovery.

> **NOTE**: `otelcol.processor.spanlogs` is a custom component unrelated 
> to any processors from the OpenTelemetry Collector. It is based on the 
> `automatic_logging` processor in the 
> [traces](../../../static/configuration/traces-config.md) 
> subsystem of the Agent static mode.

Multiple `otelcol.processor.spanlogs` components can be specified by giving them
different labels.

## Usage

```river
otelcol.processor.spanlogs "LABEL" {
  output {
    logs    = [...]
  }
}
```

## Arguments

`otelcol.processor.spanlogs` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`spans` | `bool` | Log one line per span. | `false` | no
`roots` | `bool` | Log one line for every root span of a trace. | `false` | no
`processes` | `bool` | Log one line for every process. | `false` | no
`span_attributes` | `list(string)` | Additional span attributes to log. | `[]` | no
`process_attributes` | `list(string)` | Additional process attributes to log. | `[]` | no
`labels` | `list(string)` | A lists of keys which will be logged as labels. | `[]` | no

The values listed in `labels` should be values of either span or process attributes.

> **WARNING**: Setting `spans` to `true` could lead to a high volume of logs.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.spanlogs`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
overrides | [overrides][] | Overrides for keys in the log body. | no
output | [output][] | Configures where to send received telemetry data. | yes

[output]: #output-block
[overrides]: #overrides-block

### overrides block

The `overrides` block configures overrides for keys which will be logged in the body of the log line.

The following attributes are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`logs_instance_tag` | `string` | Indicates if the log line is for a span, root or process. | `traces` | no
`service_key` | `string` | Log key for the service name of the resource. | `svc` | no
`span_name_key` | `string` | Log key for the name of the span. | `span` | no
`status_key` | `string` | Log key for the status of the span. | `status` | no
`duration_key` | `string` | Log key for the duration of the span. | `dur` | no
`trace_id_key` | `string` | Log key for the trace ID of the span. | `tid` | no

### output block

{{< docs/shared lookup="flow/reference/components/output-block-logs.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.processor.spanlogs` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.spanlogs` does not expose any component-specific debug
information.

## Example

The configuration below sends logs derived from spans to Loki.

```river
otelcol.receiver.otlp "default" {
  grpc {}

  output {
    traces = [otelcol.processor.spanlogs.default.input]
  }
}

otelcol.processor.spanlogs "default" {
    spans              = true
    roots              = true
    processes          = true
    labels             = ["attribute1", "res_attribute1"]
    span_attributes    = ["attribute1"]
    process_attributes = ["res_attribute1"]

    output {
        logs = [otelcol.exporter.loki.default.input]
    }
}

otelcol.exporter.loki "default" {
  forward_to = [loki.write.local.receiver]
}

loki.write "local" {
    endpoint {
        url = "loki:3100"
    }
```

For an input trace like this...

```json
{
    "resourceSpans": [{
        "resource": {
            "attributes": [{
                "key": "service.name",
                "value": { "stringValue": "TestSvcName" }
            },
            {
                "key": "res_attribute1",
                "value": { "intValue": "78" }
            },
            {
                "key": "unused_res_attribute1",
                "value": { "stringValue": "str" }
            },
            {
                "key": "res_account_id",
                "value": { "intValue": "2245" }
            }]
        },
        "scopeSpans": [{
            "spans": [{
                "trace_id": "7bba9f33312b3dbb8b2c2c62bb7abe2d",
                "span_id": "086e83747d0e381e",
                "name": "TestSpan",
                "attributes": [{
                    "key": "attribute1",
                    "value": { "intValue": "78" }
                },
                {
                    "key": "unused_attribute1",
                    "value": { "intValue": "78" }
                },
                {
                    "key": "account_id",
                    "value": { "intValue": "2245" }
                }]
            }]
        }]
    }]
}
```

... the output log coming out of `otelcol.processor.spanlogs` will look like this:

```json
{
    "resourceLogs": [{
        "scopeLogs": [{
            "log_records": [{
                "body": { "stringValue": "span=TestSpan dur=0ns attribute1=78 svc=TestSvcName res_attribute1=78 tid=7bba9f33312b3dbb8b2c2c62bb7abe2d" },
                "attributes": [{
                    "key": "traces",
                    "value": { "stringValue": "span" }
                },
                {
                    "key": "attribute1",
                    "value": { "intValue": "78" }
                },
                {
                    "key": "res_attribute1",
                    "value": { "intValue": "78" }
                }]
            },
            {
                "body": { "stringValue": "span=TestSpan dur=0ns attribute1=78 svc=TestSvcName res_attribute1=78 tid=7bba9f33312b3dbb8b2c2c62bb7abe2d" },
                "attributes": [{
                    "key": "traces",
                    "value": { "stringValue": "root" }
                },
                {
                    "key": "attribute1",
                    "value": { "intValue": "78" }
                },
                {
                    "key": "res_attribute1",
                    "value": { "intValue": "78" }
                }]
            },
            {
                "body": { "stringValue": "svc=TestSvcName res_attribute1=78 tid=7bba9f33312b3dbb8b2c2c62bb7abe2d" },
                "attributes": [{
                    "key": "traces",
                    "value": { "stringValue": "process" }
                },
                {
                    "key": "res_attribute1",
                    "value": { "intValue": "78" }
                }]
            }]
        }]
    }]
}
```