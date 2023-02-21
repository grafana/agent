---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.​integration.github
---

# prometheus.integration.github
The `prometheus.integration.github` component embeds
[github_exporter](https://github.com/infinityworks/github-exporter) for collecting statistics from GitHub.

## Usage

```river
prometheus.integration.github "LABEL" {
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
`users` | `list(string)` | A list of github users for which to collect metrics. | | no
`api_token`    | `string` | A github authentication token that allows the API to be queried more often. Optional, but recommended. | | no
`api_token_file`    | `string` | A path to a file containing a github authentication token that allows the API to be queried more often. If supplied, this supercedes `api_token`. Optional, but recommended.| | no

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `github` metrics.

For example, the `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metric's label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.integration.github` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.integration.github` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.integration.github` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.integration.github`:

```river
prometheus.integration.github "example" {
  api_token_file = "/etc/github-api-token"
  repositories = ["grafana/agent"]
}

// Configure a prometheus.scrape component to collect github metrics.
prometheus.scrape "demo" {
  targets    = prometheus.integration.github.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
