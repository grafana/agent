---
title: module.file
labels:
  stage: beta
---

# module.file

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`module.file` is a *module loader* component. A module loader is a Grafana Agent Flow
component which retreives a [module][] and runs the components defined inside of it.

`module.file` simplifies the configurations for modules loaded from a file by embedding
a [local.file][] component. This allows a single module loader to do the equivalence of
using the more generic [module.string][] paired with a [local.file][] component.

[module]: {{< relref "../../concepts/modules.md" >}}
[local.file]: {{< relref "./local.file.md" >}}
[module.string]: {{< relref "./module.string.md" >}}

## Usage

```river
module.file "LABEL" {
  filename  = FILENAME
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
`filename`       | `string`   | Path of the file on disk to watch | | yes
`detector`       | `string`   | Which file change detector to use (fsnotify, poll) | `"fsnotify"` | no
`poll_frequency` | `duration` | How often to poll for file changes | `"1m"` | no
`is_secret`      | `bool`     | Marks the file as containing a [secret][] | `false` | no
`arguments`      | `map(any)` | The values for the supported arguments in the module contents. | | no

`arguments` allows us to pass parameterized input into a module. The values
passed in `arguments` correspond to [argument blocks][] defined in the module
source.

An `argument` marked non-optional in the module being loaded is required in the
`arguments`. It is also not valid to provide an `argument` not defined in the
module being loaded.

### File change detectors

File change detectors are used for detecting when the file needs to be re-read
from disk. `local.file` supports two detectors: `fsnotify` and `poll`.

#### fsnotify

The `fsnotify` detector subscribes to filesystem events which indicate when the
watched file had been updated. This requires a filesystem which supports events
at the Operating System level: network-based filesystems like NFS or FUSE won't
work.

When a filesystem event is received, the component will reread the watched
file. This will happen for any filesystem event to the file, including a change
of permissions.

`fsnotify` also polls for changes to the file with the configured
`poll_frequency` as a fallback.

`fsnotify` will stop receiving filesystem events if the watched file has been
deleted, renamed, or moved. The subscription will be re-established on the next
poll once the watched file exists again.

#### poll

The `poll` file change detector will cause the watched file to be reread
every `poll_frequency`, regardless of whether the file changed.

[argument blocks]: {{< relref "../config-blocks/argument.md" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`exports` | `map(any)` | The exports of the Module loader.

`exports` exposes the `export` config block inside a module. It can be accessed
from the parent config via `module.file.LABEL.exports.EXPORT_LABEL`.

Values in `exports` correspond to [export blocks][] defined in the module
source.

[export blocks]: {{< relref "../config-blocks/export.md" >}}

## Component health

`module.file` is reported as healthy if the most recent load of the module was
successful.

If the module is not loaded successfully, the current health displays as
unhealthy and the health includes the error from loading the module.

## Debug information

`module.file` does not expose any component-specific debug information.

### Debug metrics

`module.file` does not expose any component-specific debug metrics.

## Example

In this example, we pass credentials from a parent config to a module which loads
a `prometheus.remote_write` component. The exports of the
`prometheus.remote_write` component are exposed to parent config, allowing
the parent config to pass metrics to it.

Parent:

```river
module.file "metrics" {
  filename = "/path/to/prometheus_remote_write_module.river"
  arguments = {
    username = env("PROMETHEUS_USERNAME"),
    password = env("PROMETHEUS_PASSWORD"),
  }
}

prometheus.exporter.unix { }

prometheus.scrape "local_agent" {
  targets         = prometheus.exporter.unix.targets
  forward_to      = [module.file.metrics.exports.prometheus_remote_write.receiver]
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
