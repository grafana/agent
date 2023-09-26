---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.processor.discovery/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.processor.discovery/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.processor.discovery/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.discovery/
title: otelcol.processor.discovery
description: Learn about otelcol.processor.discovery
---

# otelcol.processor.discovery

`otelcol.processor.discovery` accepts traces telemetry data from other `otelcol`
components. It can be paired with `discovery.*` components, which supply a list 
of labels for each discovered target.
`otelcol.processor.discovery` adds resource attributes to spans which have a hostname 
matching the one in the `__address__` label provided by the `discovery.*` component.

> **NOTE**: `otelcol.processor.discovery` is a custom component unrelated to any
> processors from the OpenTelemetry Collector.

Multiple `otelcol.processor.discovery` components can be specified by giving them
different labels.

## Usage

```river
otelcol.processor.discovery "LABEL" {
  targets = [...]
  output {
    traces  = [...]
  }
}
```

## Arguments

`otelcol.processor.discovery` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets` | `list(map(string))` | List of target labels to apply to the spans. | | yes
`operation_type` | `string` | Configures whether to update a span's attribute if it already exists. | `upsert` | no
`pod_associations` | `list(string)` | Configures how to decide the hostname of the span. | `["ip", "net.host.ip", "k8s.pod.ip", "hostname", "connection"]` | no

`targets` could come from `discovery.*` components:
1. The `__address__` label will be matched against the IP address of incoming spans.
   * If `__address__` contains a port, it is ignored. 
2. If a match is found, then relabeling rules are applied.
   * Note that labels starting with `__` will not be added to the spans.

The supported values for `operation_type` are:
* `insert`: Inserts a new resource attribute if the key does not already exist.
* `update`: Updates a resource attribute if the key already exists.
* `upsert`: Either inserts a new resource attribute if the key does not already exist,
   or updates a resource attribute if the key does exist.

The supported values for `pod_associations` are:
* `ip`: The hostname will be sourced from an `ip` resource attribute.
* `net.host.ip`: The hostname will be sourced from a `net.host.ip` resource attribute.
* `k8s.pod.ip`: The hostname will be sourced from a `k8s.pod.ip` resource attribute.
* `hostname`: The hostname will be sourced from a `host.name` resource attribute.
* `connection`: The hostname will be sourced from the context from the incoming requests (gRPC and HTTP).

If multiple `pod_associations` methods are enabled, the order of evaluation is honored. 
For example, when `pod_associations` is `["ip", "net.host.ip"]`, `"net.host.ip"` may be matched 
only if `"ip"` has not already matched.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.discovery`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send received telemetry data. | yes

[output]: #output-block

### output block

{{< docs/shared lookup="flow/reference/components/output-block-traces.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` OTLP-formatted data for telemetry signals of these types:
* traces

## Component health

`otelcol.processor.discovery` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.discovery` does not expose any component-specific debug
information.

## Examples

### Basic usage
```river
discovery.http "dynamic_targets" {
    url              = "https://example.com/scrape_targets"
    refresh_interval = "15s"
}

otelcol.processor.discovery "default" {
    targets = discovery.http.dynamic_targets.targets

    output {
        traces = [otelcol.exporter.otlp.default.input]
    }
}
```

### Using more than one discovery process

Outputs from more than one discovery process can be combined via the `concat` function.

```river
discovery.http "dynamic_targets" {
    url              = "https://example.com/scrape_targets"
    refresh_interval = "15s"
}

discovery.kubelet "k8s_pods" {
  bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
  namespaces        = ["default", "kube-system"]
}

otelcol.processor.discovery "default" {
    targets = concat(discovery.http.dynamic_targets.targets, discovery.kubelet.k8s_pods.targets)

    output {
        traces = [otelcol.exporter.otlp.default.input]
    }
}
```

### Using a preconfigured list of attributes

It is not necessary to use a discovery component. In the example below, a `test_label` resource 
attribute will be added to a span if its IP address is "1.2.2.2". The `__internal_label__` will
be not be added to the span, because it begins with a double underscore (`__`).

```river
otelcol.processor.discovery "default" {
    targets = [{
        "__address__"        = "1.2.2.2", 
        "__internal_label__" = "test_val",
        "test_label"         = "test_val2"}]

    output {
        traces = [otelcol.exporter.otlp.default.input]
    }
}
```
