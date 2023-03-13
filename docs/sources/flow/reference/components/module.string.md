---
title: module.string
---

# module.string

`module.string` is a *Module loader* component. A *Module loader* is a Grafana Agent Flow 
component which retreives a module and runs the components defined inside of it.

*TODO: Add link to module concept page above once merged*

## Usage

```river
module.string "LABEL" {
	content   = CONTENT
	arguments = {
		argument1 = ARGUMENT1,
		argument2 = ARGUMENT2,
		...
	}
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`content`   | `secret`    | The contents of the module to load as a secret string. | | yes
`arguments` | `map(any)`  | The values for the supported arguments in the module contents. | | no

`content` is a string that contains all the arguments, exports and components for the module. 
`content` is typically loaded via the exports of another component. For example,

- local.file.[label].content
- remote.http.[label].content
- remote.s3.[label].content

`arguments` can contain, but are not limited to strings, components and component exports.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`exports` | `map(any)` | The exports of the Module loader.

`exports` can contain, but is not limited to strings, components and component exports.

## Component health

`module.string` will be reported as healthy whenever all of its components have
been loaded successfully.

If a component is not loaded successfully, the current health will show as
unhealthy and the health will include the error from the component.

## Debug information

`module.string` does not expose any component-specific debug information.

### Debug metrics

`module.string` does not expose any component-specific debug metrics.

## Example

This example demonstrates the 3 parts (arguments, exports and components) of
a Module loader for `module.string`. In this example, we pass credentials from a
parent river config to a module for a `prometheus.remote_write` flow component.
We then export that component for use in the parent river config. The export can 
then be used by `prometheus.scrape.forward_to` to send metrics to the cloud.

Here's an example of the module `contents`. This module accepts arguments for `username` and
`password` then exports `prometheus_remote_write`. `prometheus_remote_write` is an export
of another flow component which can later be accessed by the parent river config.

```river
argument "username" { }

argument "password" { }

export "prometheus_remote_write" {
	value = prometheus.remote_write.grafana_cloud
}

prometheus.remote_write "grafana_cloud" {
	endpoint {
		url = "https://prometheus-us-central1.grafana.net/api/prom/push"

		basic_auth {
			username = argument.username.value
			password = argument.password.value
		}
	}
}
```

Here's an example of the parent river config leveraging a `local.file` component
to specify the location of the contents for `module.string`. We also access the
exported component from the module.

The `username` and `password` are set via environment variables before being passed
to the module. The `prometheus_remote_write` export from the module is accessed via
`module.string.metrics.exports.prometheus_remote_write` and then we access the
`prometheus.remote_write` export `receiver` that `prometheus.scrape` is expecting.

For the purpose of this example we have exported the entire `prometheus.remote_write.grafana_cloud`
component. Alternatively, the module above could have exported 
`prometheus.remote_write.grafana_cloud.receiver` instead of the entire component
since that is all we really need in the parent river config.

```river
local.file "metrics" {
	filename = "/path/to/prometheus_remote_write_module.river"
}

module.string "metrics" {
	content   = local.file.metrics.content
	arguments = {
		username = env("PROMETHEUS_USERNAME"),
		password = env("PROMETHEUS_PASSWORD"),
	}
}

prometheus.exporter.unix { }

prometheus.scrape "local_agent" {
	targets         = prometheus.exporter.unix.targets
	forward_to      = [module.string.metrics.exports.prometheus_remote_write.receiver]
	scrape_interval = "10s"
}
```