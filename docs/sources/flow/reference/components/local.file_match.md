---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/local.file_match/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/local.file_match/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/local.file_match/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/local.file_match/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/local.file_match/
description: Learn about local.file_match
title: local.file_match
---

# local.file_match

`local.file_match` discovers files on the local filesystem using glob patterns and the [doublestar][] library.

[doublestar]: https://github.com/bmatcuk/doublestar

## Usage

```river
local.file_match "LABEL" {
  path_targets = [{"__path__" = DOUBLESTAR_PATH}]
}
```

## Arguments

The following arguments are supported:

Name            | Type                | Description                                                                                | Default | Required
--------------- | ------------------- | ------------------------------------------------------------------------------------------ |---------| --------
`path_targets`  | `list(map(string))` | Targets to expand; looks for glob patterns on the  `__path__` and `__path_exclude__` keys. |         | yes
`sync_period`   | `duration`          | How often to sync filesystem and targets.                                                  | `"10s"` | no

`path_targets` uses [doublestar][] style paths.
* `/tmp/**/*.log` will match all subfolders of `tmp` and include any files that end in `*.log`.
* `/tmp/apache/*.log` will match only files in `/tmp/apache/` that end in `*.log`.
* `/tmp/**` will match all subfolders of `tmp`, `tmp` itself, and all files.


## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the filesystem.

Each target includes the following labels:

* `__path__`: Absolute path to the file.

## Component health

`local.file_match` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`local.file_match` does not expose any component-specific debug information.

## Debug metrics

`local.file_match` does not expose any component-specific debug metrics.

## Examples

### Send `/tmp/logs/*.log` files to Loki

This example discovers all files and folders under `/tmp/logs`. The absolute paths are 
used by `loki.source.file.files` targets.

```river
local.file_match "tmp" {
  path_targets = [{"__path__" = "/tmp/logs/**/*.log"}]
}

loki.source.file "files" {
  targets    = local.file_match.tmp.targets
  forward_to = [loki.write.endpoint.receiver]
}

loki.write "endpoint" {
  endpoint {
      url = LOKI_URL
      basic_auth {
          username = USERNAME
          password = PASSWORD
      }
  }
}
```
Replace the following:
  - `LOKI_URL`: The URL of the Loki server to send logs to.
  - `USERNAME`: The username to use for authentication to the Loki API.
  - `PASSWORD`: The password to use for authentication to the Loki API.

### Send Kubernetes pod logs to Loki

This example finds all the logs on pods and monitors them.

```river
discovery.kubernetes "k8s" {
  role = "pod"
}

discovery.relabel "k8s" {
  targets = discovery.kubernetes.k8s.targets

  rule {
    source_labels = ["__meta_kubernetes_namespace", "__meta_kubernetes_pod_label_name"]
    target_label  = "job"
    separator     = "/"
  }

  rule {
    source_labels = ["__meta_kubernetes_pod_uid", "__meta_kubernetes_pod_container_name"]
    target_label  = "__path__"
    separator     = "/"
    replacement   = "/var/log/pods/*$1/*.log"
  }
}

local.file_match "pods" {
  path_targets = discovery.relabel.k8s.output
}

loki.source.file "pods" {
  targets = local.file_match.pods.targets
  forward_to = [loki.write.endpoint.receiver]
}

loki.write "endpoint" {
  endpoint {
      url = LOKI_URL
      basic_auth {
          username = USERNAME
          password = PASSWORD
      }
  }
}
```
Replace the following:
  - `LOKI_URL`: The URL of the Loki server to send logs to.
  - `USERNAME`: The username to use for authentication to the Loki API.
  - `PASSWORD`: The password to use for authentication to the Loki API.

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`local.file_match` can accept data from the following components:

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

`local.file_match` can output data to the following components:

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
