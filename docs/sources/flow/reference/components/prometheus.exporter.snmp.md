---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.snmp/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.snmp/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.snmp/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.snmp/
title: prometheus.exporter.snmp
description: Learn about prometheus.exporter.snmp
---

# prometheus.exporter.snmp

The `prometheus.exporter.snmp` component embeds
[`snmp_exporter`](https://github.com/prometheus/snmp_exporter). `snmp_exporter` lets you collect SNMP data and expose them as Prometheus metrics.

{{% admonition type="note" %}}
`prometheus.exporter.snmp` uses the latest configuration introduced in version 0.23 of the Prometheus `snmp_exporter`.
{{% /admonition %}}

## Usage

```river
prometheus.exporter.snmp "LABEL" {
  config_file = SNMP_CONFIG_FILE_PATH

  target "TARGET_NAME" {
    address = TARGET_ADDRESS
  }
}
```

## Arguments

The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

| Name          | Type                 | Description                                      | Default | Required |
| ------------- | -------------------- | ------------------------------------------------ | ------- | -------- |
| `config_file` | `string`             | SNMP configuration file defining custom modules. |         | no       |
| `config`      | `string` or `secret` | SNMP configuration as inline string.             |         | no       |

The `config_file` argument points to a YAML file defining which snmp_exporter modules to use. See [snmp_exporter](https://github.com/prometheus/snmp_exporter#generating-configuration) for details on how to generate a config file.

The `config` argument must be a YAML document as string defining which SNMP modules and auths to use.
`config` is typically loaded by using the exports of another component. For example,

- `local.file.LABEL.content`
- `remote.http.LABEL.content`
- `remote.s3.LABEL.content`

## Blocks

The following blocks are supported inside the definition of
`prometheus.exporter.snmp` to configure collector-specific options:

| Hierarchy  | Name           | Description                                                 | Required |
| ---------- | -------------- | ----------------------------------------------------------- | -------- |
| target     | [target][]     | Configures an SNMP target.                                  | yes      |
| walk_param | [walk_param][] | SNMP connection profiles to override default SNMP settings. | no       |

[target]: #target-block
[walk_param]: #walk_param-block

### target block

The `target` block defines an individual SNMP target.
The `target` block may be specified multiple times to define multiple targets. The label of the block is required and will be used in the target's `job` label.

| Name          | Type     | Description                         | Default | Required |
| ------------- | -------- | ----------------------------------- | ------- | -------- |
| `address`     | `string` | The address of SNMP device.         |         | yes      |
| `module`      | `string` | SNMP module to use for polling.     | `""`    | no       |
| `auth`        | `string` | SNMP authentication profile to use. | `""`    | no       |
| `walk_params` | `string` | Config to use for this target.      | `""`    | no       |

### walk_param block

The `walk_param` block defines an individual SNMP connection profile that can be used to override default SNMP settings.
The `walk_param` block may be specified multiple times to define multiple SNMP connection profiles.

| Name              | Type       | Description                                   | Default | Required |
| ----------------- | ---------- | --------------------------------------------- | ------- | -------- |
| `name`            | `string`   | Name of the module to override.               |         | no       |
| `max_repetitions` | `int`      | How many objects to request with GET/GETBULK. | `25`    | no       |
| `retries`         | `int`      | How many times to retry a failed request.     | `3`     | no       |
| `timeout`         | `duration` | Timeout for each individual SNMP request.     |         | no       |

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

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
        retries = "2"
    }

    walk_param "public" {
        retries = "2"
    }
}
// Configure a prometheus.scrape component to collect SNMP metrics.
prometheus.scrape "demo" {
    targets    = prometheus.exporter.snmp.example.targets
    forward_to = [ /* ... */ ]
}
```

This example is the same above with using an embedded configuration (with secrets):

```river
local.file "snmp_config" {
    path      = "snmp_modules.yml"
    is_secret = true
}

prometheus.exporter.snmp "example" {
    config = local.file.snmp_config.content

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
        retries = "2"
    }

    walk_param "public" {
        retries = "2"
    }
}
// Configure a prometheus.scrape component to collect SNMP metrics.
prometheus.scrape "demo" {
    targets    = prometheus.exporter.snmp.example.targets
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
