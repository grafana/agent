---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.nerve/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/discovery.nerve/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.nerve/
description: Learn about discovery.nerve
title: discovery.nerve
---

# discovery.nerve

`discovery.nerve` discovers [airbnb/nerve][] targets stored in Zookeeper.

[airbnb/nerve]: https://github.com/airbnb/nerve

## Usage

```river
discovery.nerve "LABEL" {
	servers = [SERVER_1, SERVER_2]
	paths   = [PATH_1, PATH_2]
}
```

## Arguments

The following arguments are supported:

Name               | Type           | Description                          | Default       | Required
------------------ | -------------- | ------------------------------------ | ------------- | --------
`servers`          | `list(string)` | The Zookeeper servers.               |               | yes
`paths`            | `list(string)` | The paths to look for targets at.    |               | yes
`timeout`          | `duration`     | The timeout to use.                  | `"10s"`       | no


Each element in the `path` list can either point to a single service, or to the
root of a tree of services.

## Blocks

The `discovery.nerve` component does not support any blocks, and is configured
fully through arguments.

## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from Nerve's API.

The following meta labels are available on targets and can be used by the
discovery.relabel component
* `__meta_nerve_path`: the full path to the endpoint node in Zookeeper
* `__meta_nerve_endpoint_host`: the host of the endpoint
* `__meta_nerve_endpoint_port`: the port of the endpoint
* `__meta_nerve_endpoint_name`: the name of the endpoint

## Component health

`discovery.nerve` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.nerve` does not expose any component-specific debug information.

## Debug metrics

`discovery.nerve` does not expose any component-specific debug metrics.

## Example

```river
discovery.nerve "example" {
	servers = ["localhost"]
	paths   = ["/monitoring"]
	timeout = "1m"
}
prometheus.scrape "demo" {
	targets    = discovery.nerve.example.targets
	forward_to = [prometheus.remote_write.demo.receiver]
}
prometheus.remote_write "demo" {
	endpoint {
		url = PROMETHEUS_REMOTE_WRITE_URL
		basic_auth {
			username = USERNAME
			password = PASSWORD
		}
	}
}
```

Replace the following:
  - `PROMETHEUS_REMOTE_WRITE_URL`: The URL of the Prometheus remote_write-compatible server to send metrics to.
  - `USERNAME`: The username to use for authentication to the remote_write API.
  - `PASSWORD`: The password to use for authentication to the remote_write API.

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`discovery.nerve` can output data to the following components:

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
