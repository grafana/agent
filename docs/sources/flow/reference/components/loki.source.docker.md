---
aliases:
- /docs/agent/latest/flow/reference/components/loki.source.docker/
- /docs/grafana-cloud/agent/flow/reference/components/loki.source.docker/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/loki.source.docker/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/loki.source.docker/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.source.docker/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.docker/
description: Learn about loki.source.docker
title: loki.source.docker
---

# loki.source.docker

`loki.source.docker` reads log entries from Docker containers and forwards them
to other `loki.*` components. Each component can read from a single Docker
daemon.

Multiple `loki.source.docker` components can be specified by giving them
different labels.

## Usage

```river
loki.source.docker "LABEL" {
  host       = HOST
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments
The component starts a new reader for each of the given `targets` and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.file` supports the following arguments:

Name            | Type                 | Description          | Default | Required
--------------- | -------------------- | -------------------- | ------- | --------
`host`          | `string`             | Address of the Docker daemon. | | yes
`targets`       | `list(map(string))`  | List of containers to read logs from. | | yes
`forward_to`    | `list(LogsReceiver)` | List of receivers to send log entries to. | | yes
`labels`        | `map(string)`        | The default set of labels to apply on entries. | `"{}"` | no
`relabel_rules` | `RelabelRules`       | Relabeling rules to apply on log entries. | `"{}"` | no
`refresh_interval` | `duration`        | The refresh interval to use when connecting to the Docker daemon over HTTP(S). | `"60s"` | no

## Blocks

The following blocks are supported inside the definition of `loki.source.docker`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client | [client][] | HTTP client settings when connecting to the endpoint. | no
client > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
client > authorization | [authorization][] | Configure generic authorization to the endpoint. | no
client > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
client > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no
client > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example, `client >
basic_auth` refers to an `basic_auth` block defined inside a `client` block.

These blocks are only applicable when connecting to a Docker daemon over HTTP
or HTTPS and has no effect when connecting via a `unix:///` socket

[client]: #client-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### client block

The `client` block configures settings used to connect to HTTP(S) Docker
daemons.

{{< docs/shared lookup="flow/reference/components/http-client-config-block.md" source="agent" version="<AGENT_VERSION>" >}}

### basic_auth block

The `basic_auth` block configures basic authentication for HTTP(S) Docker
daemons.

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT_VERSION>" >}}

### authorization block

The `authorization` block configures custom authorization to use for the Docker
daemon.

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT_VERSION>" >}}

### oauth2 block

The `oauth2` block configures OAuth2 authorization to use for the Docker
daemon.

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT_VERSION>" >}}

### tls_config block

The `tls_config` block configures TLS settings for connecting to HTTPS Docker
daemons.

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT_VERSION>" >}}

## Exported fields

`loki.source.docker` does not export any fields.

## Component health

`loki.source.docker` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.docker` exposes some debug information per target:
* Whether the target is ready to tail entries.
* The labels associated with the target.
* The most recent time a log line was read.

## Debug metrics

* `loki_source_docker_target_entries_total` (gauge): Total number of successful entries sent to the Docker target.
* `loki_source_docker_target_parsing_errors_total` (gauge): Total number of parsing errors while receiving Docker messages.

## Component behavior
The component uses its data path (a directory named after the domain's
fully qualified name) to store its _positions file_. The positions file
stores the read offsets so that if there is a component or Agent restart,
`loki.source.docker` can pick up tailing from the same spot.

## Example

This example collects log entries from the files specified in the `targets`
argument and forwards them to a `loki.write` component to be written to Loki.

```river
discovery.docker "linux" {
  host = "unix:///var/run/docker.sock"
}

loki.source.docker "default" {
  host       = "unix:///var/run/docker.sock"
  targets    = discovery.docker.linux.targets 
  forward_to = [loki.write.local.receiver]
}

loki.write "local" {
  endpoint {
    url = "loki:3100/api/v1/push"
  }
}
```

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`loki.source.docker` can accept data from the following components:

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

`loki.source.docker` can output data to the following components:

- Components that accept Loki Logs:
  - [`loki.echo`]({{< relref "../components/loki.echo.md" >}})
  - [`loki.process`]({{< relref "../components/loki.process.md" >}})
  - [`loki.relabel`]({{< relref "../components/loki.relabel.md" >}})
  - [`loki.write`]({{< relref "../components/loki.write.md" >}})
  - [`otelcol.receiver.loki`]({{< relref "../components/otelcol.receiver.loki.md" >}})

Note that connecting some components may not be feasible or components may require further configuration to make the connection work correctly. Please refer to the linked documentation for more details.

<!-- END GENERATED COMPATIBLE COMPONENTS -->
