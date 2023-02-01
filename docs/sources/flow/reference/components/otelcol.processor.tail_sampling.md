---
title: otelcol.​processor.tail_sampling
---

# otelcol.processor.tail_sampling

`otelcol.processor.tail_sampling` samples traces based on a set of defined 
policies. All spans for a given trace MUST be received by the same collector
instance for effective sampling decisions.

The `tail_sampling` component uses both soft and hard limits, where the hard limit
is always equal or larger than the soft limit. When memory usage goes above the
soft limit, the processor component drops data and returns errors to the
preceding components in the pipeline. When usage exceeds the hard
limit, the processor forces a garbage collection in order to try and free
memory. When usage is below the soft limit, no data is dropped and no forced
garbage collection is performed.

> **NOTE**: `otelcol.processor.tail_sampling` is a wrapper over the upstream
> OpenTelemetry Collector Contrib `tail_sampling` processor. Bug reports or feature
> requests will be redirected to the upstream repository, if necessary.

Multiple `otelcol.processor.tail_sampling` components can be specified by
giving them different labels.

## Usage

```river
otelcol.processor.tail_sampling "LABEL" {
  decision_wait               = "10s"
  num_traces                  = 100
  expected_new_traces_per_sec = 10
  policy {
    ...
  }
  policy {
    ...
  }
  ...

  output {
    traces  = [...]
  }
}
```

## Arguments

`otelcol.processor.tail_sampling` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`decision_wait`               | `duration` | Wait time since the first span of a trace before making a sampling decision. | `"30s"` | no
`num_traces`                  | `int`      | Number of traces kept in memory. | `50000` | no
`expected_new_traces_per_sec` | `int`      | Expected number of new traces (helps in allocating data structures). | `0` | no

The arguments must define either `limit` or the `limit_percentage,
spike_limit_percentage` pair, but not both.

The configuration options `limit` and `limit_percentage` define the hard
limits. The soft limits are then calculated as the hard limit minus the
`spike_limit` or `spike_limit_percentage` values respectively. The recommended
value for spike limits is about 20% of the corresponding hard limit.

The recommended `check_interval` value is 1 second. If the traffic through the
component is spiky in nature, it is recommended to either decrease the interval
or increase the spike limit to avoid going over the hard limit.

The `limit` and `spike_limit` values must be larger than 1 MiB.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.tail_sampling`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
policy             | [policy][] | Policies used to make a sampling decision. | yes
policy > and       | [and] | Sample based on multiple policies, creates an AND policy | no
policy > composite | [composite] | Sample based on a combination of samplers.  | no
output             | [output][] | Configures where to send received telemetry data. | yes

[policy]: #policy-block

### policy block

POLICY INFO

[and]: #and-block

### and block

AND INFO

[composite]: #composite-block

### composite block

COMPOSITE INFO

[output]: #output-block

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.processor.tail_sampling` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.tail_sampling` does not expose any component-specific debug
information.

## Example

This example batches trace data from the agent before sending it to
[otelcol.exporter.otlp][] for further processing. It shows an impracticle amount of policies for the purpose of demonstrating how to set each type up:

```river
tracing {
	sampling_fraction = 1
	write_to          = [otelcol.processor.tail_sampling.default.input]
}

otelcol.processor.tail_sampling "default" {
  decision_wait               = "10s"
  num_traces                  = 100
  expected_new_traces_per_sec = 10
  policy {
    name = "test-policy-1"
    type = "always_sample"
  }
  policy {
    name = "test-policy-2"
  	type = "latency"
    latency {
      threshold_ms = 5000
    }
  }
  policy {
    name = "test-policy-3"
    type = "numeric_attribute"
    numeric_attribute {
      key       = "key1"
      min_value = 50
      max_value = 100
    }
  }
  policy {
    name = "test-policy-4"
    type = "probabilistic"
    probabilistic {
      sampling_percentage = 10
    }
  }
  policy {
    name = "test-policy-5"
    type = "status_code"
    status_code {
      status_codes = ["ERROR", "UNSET"]
    }
  }
  policy {
    name = "test-policy-6"
    type = "string_attribute"
    string_attribute {
      key    = "key2"
      values = ["value1", "value2"]
    }
  }
  policy {
    name = "test-policy-7"
    type = "string_attribute"
    string_attribute {
      key                    = "key2"
      values                 = ["value1", "val*"]
      enabled_regex_matching = true
      cache_max_size         = 10
    }
  }
  policy {
    name = "test-policy-8"
    type = "rate_limiting"
    rate_limiting {
      spans_per_second = 35
    }
  }
  policy {
    name = "test-policy-9"
    type = "string_attribute"
    string_attribute {
      key                    = "http.url"
      values                 = ["/health", "/metrics"]
      enabled_regex_matching = true
      invert_match           = true
    }
  }
  policy {
    name = "test-policy-10"
    type = "span_count"
    span_count {
      min_spans = 2
    }
  }
  policy {
    name = "test-policy-11"
    type = "trace_state"
    trace_state {
      key    = "key3"
      values = ["value1", "value2"]
    }
  }
  policy{
    name = "and-policy-1"
    type = "and"
    and {
      and_sub_policy {
        name = "test-and-policy-1"
        type = "numeric_attribute"
        numeric_attribute {
          key       = "key1"
          min_value = 50
          max_value = 100
        }
      }
      and_sub_policy {
        name = "test-and-policy-2"
        type = "string_attribute"
        string_attribute {
          key    = "key1"
          values = ["value1", "value2"]
        }
      }
    }
  }
  policy{
    name = "composite-policy-1"
    type = "composite"
    composite {
      max_total_spans_per_second = 1000
      policy_order               = ["test-composite-policy-1", "test-composite-policy-2", "test-composite-policy-3"]
      composite_sub_policy {
        name = "test-composite-policy-1"
        type = "numeric_attribute"
        numeric_attribute {
          key       = "key1"
          min_value = 50
          max_value = 100
        }
      }
      composite_sub_policy {
        name = "test-composite-policy-2"
        type = "string_attribute"
        string_attribute {
          key    = "key1"
          values = ["value1", "value2"]
        }
      }
      composite_sub_policy {
        name = "test-composite-policy-3"
        type = "always_sample"
      }
      rate_allocation {
        policy  = "test-composite-policy-1"
        percent = 50
      }
      rate_allocation {
        policy  = "test-composite-policy-2"
        percent = 50
      }
    }
  }
  
  output {
	  traces = [otelcol.exporter.otlp.production.input]
  }
}

otelcol.exporter.otlp "production" {
  client {
    endpoint = env("OTLP_SERVER_ENDPOINT")
  }
}
```