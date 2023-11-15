---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.relabel/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.relabel/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.relabel/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/discovery.relabel/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.relabel/
description: Learn about discovery.relabel
title: discovery.relabel
---

# discovery.relabel

In Flow, targets are defined as sets of key-value pairs called _labels_.

`discovery.relabel` rewrites the label set of the input targets by applying one
or more relabeling rules. If no rules are defined, then the input targets are
exported as-is.

The most common use of `discovery.relabel` is to filter targets or standardize
the target label set that is passed to a downstream component. The `rule`
blocks are applied to the label set of each target in order of their appearance
in the configuration file. The configured rules can be retrieved by calling the
function in the `rules` export field.

Target labels which start with a double underscore `__` are considered
internal, and may be removed by other Flow components prior to telemetry
collection. To retain any of these labels, use a `labelmap` action to remove
the prefix, or remap them to a different name. Service discovery mechanisms
usually group their labels under `__meta_*`. For example, the
discovery.kubernetes component populates a set of `__meta_kubernetes_*` labels
to provide information about the discovered Kubernetes resources. If a
relabeling rule needs to store a label value temporarily, for example as the
input to a subsequent step, use the `__tmp` label name prefix, as it is
guaranteed to never be used.

Multiple `discovery.relabel` components can be specified by giving them
different labels.

## Usage

```river
discovery.relabel "LABEL" {
  targets = TARGET_LIST

  rule {
    ...
  }

  ...
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets` | `list(map(string))` | Targets to relabel | | yes

## Blocks

The following blocks are supported inside the definition of
`discovery.relabel`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
rule | [rule][] | Relabeling rules to apply to targets. | no

[rule]: #rule-block

### rule block

{{< docs/shared lookup="flow/reference/components/rule-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`output` | `list(map(string))` | The set of targets after applying relabeling.
`rules`    | `RelabelRules` | The currently configured relabeling rules.

## Component health

`discovery.relabel` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.relabel` does not expose any component-specific debug information.

## Debug metrics

`discovery.relabel` does not expose any component-specific debug metrics.

## Example

```river
discovery.relabel "keep_backend_only" {
  targets = [
    { "__meta_foo" = "foo", "__address__" = "localhost", "instance" = "one",   "app" = "backend"  },
    { "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "two",   "app" = "database" },
    { "__meta_baz" = "baz", "__address__" = "localhost", "instance" = "three", "app" = "frontend" },
  ]

  rule {
    source_labels = ["__address__", "instance"]
    separator     = "/"
    target_label  = "destination"
    action        = "replace"
  }

  rule {
    source_labels = ["app"]
    action        = "keep"
    regex         = "backend"
  }
}
```


<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`discovery.relabel` can accept data from the following components:

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

`discovery.relabel` can output data to the following components:

- Components that accept Targets:
  - [`discovery.relabel`]({{< relref "../components/discovery.relabel.md" >}})
  - [`local.file_match`]({{< relref "../components/local.file_match.md" >}})
  - [`loki.source.docker`]({{< relref "../components/loki.source.docker.md" >}})
  - [`loki.source.file`]({{< relref "../components/loki.source.file.md" >}})
  - [`loki.source.kubernetes`]({{< relref "../components/loki.source.kubernetes.md" >}})
  - [`otelcol.processor.discovery`]({{< relref "../components/otelcol.processor.discovery.md" >}})
  - [`prometheus.scrape`]({{< relref "../components/prometheus.scrape.md" >}})
  - [`pyroscope.scrape`]({{< relref "../components/pyroscope.scrape.md" >}})

Note that connecting some components may not be feasible or components may require further configuration to make the connection work correctly. Please refer to the linked documentation for more details.

<!-- END GENERATED COMPATIBLE COMPONENTS -->
