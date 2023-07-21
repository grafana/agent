---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.github/
title: prometheus.exporter.github
---

# prometheus.exporter.github
The `prometheus.exporter.github` component embeds
[github_exporter](https://github.com/infinityworks/github-exporter) for collecting statistics from GitHub.

## Usage

```river
prometheus.exporter.github "LABEL" {
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`api_url`    | `string` | The full URI of the GitHub API. | `https://api.github.com` | no
`repositories` | `list(string)` | GitHub repositories for which to collect metrics. | | no
`organizations` | `list(string)` | GitHub organizations for which to collect metrics. | | no
`users` | `list(string)` | A list of GitHub users for which to collect metrics. | | no
`api_token`    | `secret` | API token to use to authenticate against GitHub. | | no
`api_token_file`    | `string` | File containing API token to use to authenticate against GitHub. | | no

GitHub uses an aggressive rate limit for unauthenticated requests based on IP address. To allow more API requests, it is recommended to configure either `api_token` or `api_token_file` to authenticate against GitHub.

When provided, `api_token_file` takes precedence over `api_token`.

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `github` metrics.

For example, the `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metric's label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.github` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.github` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.github` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.github`:

```river
prometheus.exporter.github "example" {
  api_token_file = "/etc/github-api-token"
  repositories   = ["grafana/agent"]
}

// Configure a prometheus.scrape component to collect github metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.github.example.targets
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

[scrape]: {{< relref "./prometheus.scrape.md" >}}
