---
title: module.string
---

# module.string

`module.string` is a *module loader* component. A module loader is a Grafana Agent Flow 
component which retreives a module and runs the components defined inside of it.

*TODO: Add link to modules concept page once merged*

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
`content`   | `secret` or `string` | The contents of the module to load as a secret or string. | | yes
`arguments` | `map(any)`  | The values for the supported arguments in the module contents. | | no

`content` is a string that contains the configuration of the module to load.
`content` is typically loaded by using the exports of another component. For example,

- `local.file.LABEL.content`
- `remote.http.LABEL.content`
- `remote.s3.LABEL.content`

`arguments` allows us to pass parameterized input into a module.
An `argument` marked non-optional in the module being loaded is required in the
`arguments`. It is also not valid to provide an `argument` not defined in the
module being loaded.

*TODO: Add link to argument config-blocks page once merged*

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`exports` | `map(any)` | The exports of the Module loader.

`exports` exposes the `export` config block inside a module. It can be accessed from
the parent config via `module.string.LABEL.exports.EXPORT_LABEL

*TODO: Add link to export config-blocks page once merged*

## Component health

`module.string` is reported as healthy if the most recent load of the module was 
successful. 

If the module is not loaded successfully, the current health displays as
unhealthy and the health includes the error from loading the module.

## Debug information

`module.string` does not expose any component-specific debug information.

### Debug metrics

`module.string` does not expose any component-specific debug metrics.

## Example

In this example, we pass credentials from a parent config to a module which loads
a `prometheus.remote_write` component. The exports of the 
`prometheus.remote_write` component are exposed to parent config, allowing 
the parent config to pass metrics to it.

Parent:

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

Module:

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
