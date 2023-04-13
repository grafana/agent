---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.blackbox
---

# prometheus.exporter.blackbox
The `prometheus.exporter.blackbox` component embeds
[`blackbox_exporter`](https://github.com/prometheus/blackbox_exporter). `blackbox_exporter` lets you collect blackbox metrics (probes) and expose them as Prometheus metrics.

## Usage

```river
prometheus.exporter.blackbox "LABEL" {
  config_file = "PATH_BLACKBOX_CONFIG_FILE"
  
  target "TARGET_NAME" {
    address = "TARGET_ADDRESS" 
  }
  
  ...
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`config_file`                 | `string`       | Blackbox configuration file with custom modules. | | no
`config`                      | `string`       | Blackbox configuration with custom modules as YAML. | |no
`probe_timeout_offset`        | `duration`     | Offset in seconds to subtract from timeout when probing targets.  | `"0.5s"` | no

The `config_file` argument points to a YAML file defining which blackbox_exporter modules to use. See [blackbox_exporter]( https://github.com/prometheus/blackbox_exporter/blob/master/example.yml) for details on how to generate a config file.

## Blocks

The following blocks are supported inside the definition of
`prometheus.exporter.blackbox` to configure collector-specific options:

Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
target | [target][] | Configures a blackbox target. | yes

[target]: #target-block

### target block

The `target` block defines an individual blackbox target.
The `target` block may be specified multiple times to define multiple targets.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name` | `string` | Name of the target. | | yes
`address` | `string` | The address of the target to probe. | | yes
`module`| `string` | Blackbox module to use to probe. | `""` | no

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `blackbox` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.exporter.blackbox` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.blackbox` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.blackbox` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.blackbox`:

```river
prometheus.exporter.blackbox "example" { 
	config_file = "blackbox_modules.yml"
	
	target "example" {
		address = "http://example.com"
		module  = "http_2xx"
	}
	
	target "grafana" {
		address = "http://grafana.com"
		module  = "http_2xx"
	}	
}

// Configure a prometheus.scrape component to collect Blackbox metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.blackbox.example.targets
  forward_to = [ /* ... */ ]
}
```

This example is the same above with using an embedded configuration:

```river
prometheus.exporter.blackbox "example" { 
	config = "{ modules: { http_2xx: { prober: http, timeout: 5s } } }"
	
	target "example" {
		address = "http://example.com"
		module  = "http_2xx"
	}
	
	target "grafana" {
		address = "http://grafana.com"
		module  = "http_2xx"
	}	
}

// Configure a prometheus.scrape component to collect Blackbox metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.blackbox.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
