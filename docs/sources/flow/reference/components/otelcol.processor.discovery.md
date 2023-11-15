---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.processor.discovery/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.processor.discovery/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.processor.discovery/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.processor.discovery/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.discovery/
description: Learn about otelcol.processor.discovery
title: otelcol.processor.discovery
---

# otelcol.processor.discovery

`otelcol.processor.discovery` accepts traces telemetry data from other `otelcol`
components. It can be paired with `discovery.*` components, which supply a list 
of labels for each discovered target.
`otelcol.processor.discovery` adds resource attributes to spans which have a hostname 
matching the one in the `__address__` label provided by the `discovery.*` component.

{{% admonition type="note" %}}
`otelcol.processor.discovery` is a custom component unrelated to any
processors from the OpenTelemetry Collector.
{{% /admonition %}}

Multiple `otelcol.processor.discovery` components can be specified by giving them
different labels.

{{% admonition type="note" %}}
It can be difficult to follow [OpenTelemetry semantic conventions][OTEL sem conv] when 
adding resource attributes via `otelcol.processor.discovery`:
* `discovery.relabel` and most `discovery.*` processes such as `discovery.kubernetes` 
  can only emit [Prometheus-compatible labels][Prometheus data model].
* Prometheus labels use underscores (`_`) in labels names, whereas 
  [OpenTelemetry semantic conventions][OTEL sem conv] use dots (`.`).
* Although `otelcol.processor.discovery` is able to work with non-Prometheus labels
  such as ones containing dots, the fact that `discovery.*` components are generally 
  only compatible with Prometheus naming conventions makes it hard to follow OpenTelemetry 
  semantic conventions in `otelcol.processor.discovery`.

If your use case is to add resource attributes which contain Kubernetes metadata, 
consider using `otelcol.processor.k8sattributes` instead.

------
The main use case for `otelcol.processor.discovery` is for users who migrate to Grafana Agent Flow mode
from Static mode's `prom_sd_operation_type`/`prom_sd_pod_associations` [configuration options][Traces].

[Prometheus data model]: https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
[OTEL sem conv]: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/README.md
[Traces]: {{< relref "../../../static/configuration/traces-config.md" >}}
{{% /admonition %}}

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

It is not necessary to use a discovery component. In the example below, both a `test_label` and 
a `test.label.with.dots` resource attributes will be added to a span if its IP address is 
"1.2.2.2". The `__internal_label__` will be not be added to the span, because it begins with 
a double underscore (`__`).

```river
otelcol.processor.discovery "default" {
    targets = [{
        "__address__"          = "1.2.2.2", 
        "__internal_label__"   = "test_val",
        "test_label"           = "test_val2",
        "test.label.with.dots" = "test.val2.with.dots"}]

    output {
        traces = [otelcol.exporter.otlp.default.input]
    }
}
```

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`otelcol.processor.discovery` can accept data from the following components:

- Components that output Targets:
  - [`discovery.azure`]({{< relref "../components/discovery.azure.md" >}})
  - [`discovery.consul`]({{< relref "../components/discovery.consul.md" >}})
  - [`discovery.consulagent`]({{< relref "../components/discovery.consulagent.md" >}})
  - [`discovery.digitalocean`]({{< relref "../components/discovery.digitalocean.md" >}})
  - [`discovery.dns`]({{< relref "../components/discovery.dns.md" >}})
  - [`discovery.docker`]({{< relref "../components/discovery.docker.md" >}})
  - [`discovery.dockerswarm`]({{< relref "../components/discovery.dockerswarm.md" >}})
  - [`discovery.ec2`]({{< relref "../components/discovery.ec2.md" >}})
  - [`discovery.eureka`]({{< relref "../components/discovery.eureka.md" >}})
  - [`discovery.file`]({{< relref "../components/discovery.file.md" >}})
  - [`discovery.gce`]({{< relref "../components/discovery.gce.md" >}})
  - [`discovery.hetzner`]({{< relref "../components/discovery.hetzner.md" >}})
  - [`discovery.http`]({{< relref "../components/discovery.http.md" >}})
  - [`discovery.ionos`]({{< relref "../components/discovery.ionos.md" >}})
  - [`discovery.kubelet`]({{< relref "../components/discovery.kubelet.md" >}})
  - [`discovery.kubernetes`]({{< relref "../components/discovery.kubernetes.md" >}})
  - [`discovery.kuma`]({{< relref "../components/discovery.kuma.md" >}})
  - [`discovery.lightsail`]({{< relref "../components/discovery.lightsail.md" >}})
  - [`discovery.linode`]({{< relref "../components/discovery.linode.md" >}})
  - [`discovery.marathon`]({{< relref "../components/discovery.marathon.md" >}})
  - [`discovery.nerve`]({{< relref "../components/discovery.nerve.md" >}})
  - [`discovery.nomad`]({{< relref "../components/discovery.nomad.md" >}})
  - [`discovery.openstack`]({{< relref "../components/discovery.openstack.md" >}})
  - [`discovery.puppetdb`]({{< relref "../components/discovery.puppetdb.md" >}})
  - [`discovery.relabel`]({{< relref "../components/discovery.relabel.md" >}})
  - [`discovery.scaleway`]({{< relref "../components/discovery.scaleway.md" >}})
  - [`discovery.serverset`]({{< relref "../components/discovery.serverset.md" >}})
  - [`discovery.triton`]({{< relref "../components/discovery.triton.md" >}})
  - [`discovery.uyuni`]({{< relref "../components/discovery.uyuni.md" >}})
  - [`local.file_match`]({{< relref "../components/local.file_match.md" >}})
  - [`prometheus.exporter.agent`]({{< relref "../components/prometheus.exporter.agent.md" >}})
  - [`prometheus.exporter.apache`]({{< relref "../components/prometheus.exporter.apache.md" >}})
  - [`prometheus.exporter.azure`]({{< relref "../components/prometheus.exporter.azure.md" >}})
  - [`prometheus.exporter.blackbox`]({{< relref "../components/prometheus.exporter.blackbox.md" >}})
  - [`prometheus.exporter.cadvisor`]({{< relref "../components/prometheus.exporter.cadvisor.md" >}})
  - [`prometheus.exporter.cloudwatch`]({{< relref "../components/prometheus.exporter.cloudwatch.md" >}})
  - [`prometheus.exporter.consul`]({{< relref "../components/prometheus.exporter.consul.md" >}})
  - [`prometheus.exporter.dnsmasq`]({{< relref "../components/prometheus.exporter.dnsmasq.md" >}})
  - [`prometheus.exporter.elasticsearch`]({{< relref "../components/prometheus.exporter.elasticsearch.md" >}})
  - [`prometheus.exporter.gcp`]({{< relref "../components/prometheus.exporter.gcp.md" >}})
  - [`prometheus.exporter.github`]({{< relref "../components/prometheus.exporter.github.md" >}})
  - [`prometheus.exporter.kafka`]({{< relref "../components/prometheus.exporter.kafka.md" >}})
  - [`prometheus.exporter.memcached`]({{< relref "../components/prometheus.exporter.memcached.md" >}})
  - [`prometheus.exporter.mongodb`]({{< relref "../components/prometheus.exporter.mongodb.md" >}})
  - [`prometheus.exporter.mssql`]({{< relref "../components/prometheus.exporter.mssql.md" >}})
  - [`prometheus.exporter.mysql`]({{< relref "../components/prometheus.exporter.mysql.md" >}})
  - [`prometheus.exporter.oracledb`]({{< relref "../components/prometheus.exporter.oracledb.md" >}})
  - [`prometheus.exporter.postgres`]({{< relref "../components/prometheus.exporter.postgres.md" >}})
  - [`prometheus.exporter.process`]({{< relref "../components/prometheus.exporter.process.md" >}})
  - [`prometheus.exporter.redis`]({{< relref "../components/prometheus.exporter.redis.md" >}})
  - [`prometheus.exporter.snmp`]({{< relref "../components/prometheus.exporter.snmp.md" >}})
  - [`prometheus.exporter.snowflake`]({{< relref "../components/prometheus.exporter.snowflake.md" >}})
  - [`prometheus.exporter.squid`]({{< relref "../components/prometheus.exporter.squid.md" >}})
  - [`prometheus.exporter.statsd`]({{< relref "../components/prometheus.exporter.statsd.md" >}})
  - [`prometheus.exporter.unix`]({{< relref "../components/prometheus.exporter.unix.md" >}})
  - [`prometheus.exporter.vsphere`]({{< relref "../components/prometheus.exporter.vsphere.md" >}})
  - [`prometheus.exporter.windows`]({{< relref "../components/prometheus.exporter.windows.md" >}})


Note that connecting some components may not be feasible or components may require further configuration to make the connection work correctly. Please refer to the linked documentation for more details.

<!-- END GENERATED COMPATIBLE COMPONENTS -->
