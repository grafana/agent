---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.snmp
---

# prometheus.exporter.snmp
The `prometheus.exporter.snmp` component embeds
[`snmp_exporter`](https://github.com/prometheus/snmp_exporter). `snmp_exporter` lets you collect SNMP data and expose them as Prometheus metrics.

## Usage

```river
prometheus.exporter.snmp "LABEL" {
  config_file = "PATH_SNMP_CONFIG_FILE"

  target "TARGET_NAME" {
    address = "TARGET_ADDRESS"
  }

  walk_param "PARAM_NAME" {
  }

  ...
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`config_file` | `string`       | SNMP configuration file defining custom modules. | | yes

The `config_file` argument points to a YAML file defining which snmp_exporter modules to use. See [snmp_exporter](https://github.com/prometheus/snmp_exporter#generating-configuration) for details on how to generate a config file.

## Blocks

The following blocks are supported inside the definition of
`prometheus.exporter.snmp` to configure collector-specific options:

Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
target | [target][] | Configures an SNMP target. | yes
walk_param | [walk_param][] | SNMP connection profiles to override default SNMP settings. | no
walk_param > auth | [auth][] | Configure auth for authenticating to the endpoint. | no

[target]: #target-block
[walk_param]: #walk_param-block
[auth]: #auth-block

### target block

The `target` block defines an individual SNMP target.
The `target` block may be specified multiple times to define multiple targets.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name` | `string` | Name of a snmp_target. | | yes
`address` | `string` | The address of SNMP device. | | yes
`module`| `string` | SNMP module to use for polling. | `""` | no
`walk_params`| `string` | Config to use for this target. | `""` | no

### walk_param block

The `walk_param` block defines an individual SNMP connection profile that can be used to override default SNMP settings.
The `walk_param` block may be specified multiple times to define multiple SNMP connection profiles.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name` | `string` | Name of the module to override. | | no
`version` | `int` | SNMP version to use | `2` | no
`max_repetitions`| `int` | How many objects to request with GET/GETBULK. | `25` | no
`retries`| `int` | How many times to retry a failed request. | `3` | no
`timeout`| `duration` | Timeout for each individual SNMP request. |  | no
`auth` | [auth][] | Configure auth for walk param. | | no

`version` 1 will use GETNEXT, 2 and 3 use GETBULK.

### auth block

The `auth` block defines an individual SNMP authentication profile that can be used to override default SNMP auth settings.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`community` | `secret` | Community string is used with SNMP v1 and v2. | `"public"` | no
`username` | `string` | NetSNMP username. | `"user"` | no
`security_level` | `string` | NetSNMP security_level. | `noAuthNoPriv`| no
`password` | `secret` | NetSNMP password. | `""` | no
`auth_protocol` | `string` | NetSNMP auth protocol. | `"MD5"` | no
`priv_protocol` | `string` | NetSNMP privacy protocol. | `"DES"` | no
`priv_password` | `secret` | NetSNMP privacy password. | `""` | no
`context_name` | `string` | NetSNMP context name. | `""`| no

`username` is required if v3 is used. `-u option` to NetSNMP.
`security_level` can be `noAuthNoPriv`, `authNoPriv` or `authPriv`. `-l option` to NetSNMP.
`password` is also known as `authKey`. Is required if `security_level` is `authNoPriv` or `authPriv`. `-a option` to NetSNMP.
`auth_protocol` is used if `security_level` is `authNoPriv` or `authPriv`. Possible values are `MD5`, `SHA`, `SHA224`, `SHA256`, `SHA384`, or `SHA512`. `-a option` to NetSNMP.
`priv_protocol` is used if `security_level` is `authPriv`. Possible values are `DES`, `AES`, `AES192`, or `AES256`. `-x option` to NetSNMP.
`priv_password` is also known as `privKey`. Is required if `security_level` is `authPriv`. `-x option` to NetSNMP.
`context_name` is required if context is configured on the device. `-n option` to NetSNMP.

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `snmp` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.snmp` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.snmp` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.snmp` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.snmp`:

```river
prometheus.exporter.snmp "example" {
    config_file = "snmp_modules.yml"

    target "network_switch_1" {
        address     = "192.168.1.2"
        module      = "if_mib"
        walk_params = "public"
    }

    target "network_router_2" {
        address     = "192.168.1.3"
        module      = "mikrotik"
        walk_params = "private"
    }

    walk_param "private" {
        version = "2"

        auth {
            community = "secret"
        }
    }

    walk_param "public" {
        version = "2"

        auth {
            community = "public"
        }
    }
}
// Configure a prometheus.scrape component to collect SNMP metrics.
prometheus.scrape "demo" {
    targets    = prometheus.exporter.snmp.example.targets
    forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
