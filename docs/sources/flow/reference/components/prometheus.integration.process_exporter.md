---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.​integration.​process_exporter
---

# prometheus.integration.process_exporter
The `prometheus.integration.process_exporter` component embeds
[process_exporter](https://github.com/ncabatoff/process-exporter) for collecting process stats mined from `/proc`.

## Usage

```river
prometheus.integration.process_exporter "LABEL" {
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`procfs_path`       | `string`                 | procfs mountpoint. | `/proc` | no
`track_children`    | `bool`                   | If a proc is tracked, also track any of it's children that are not part of their own group. | `true` | no
`track_threads`     | `bool`                   | Report on per-threadname metrics as well. | `true` | no
`gather_smaps`      | `bool`                   | Gather metrics from the smaps file, which contains proportional resident memory size. | `true` | no
`recheck_on_scrape` | `bool`                   | Recheck process names on each scrape. | `true` | no

## Blocks
The following blocks are supported inside the definition of `prometheus.integration.process_exporter`:

Hierarchy        | Block      | Description | Required
---------------- | ---------- | ----------- | --------
process_names          | [process_names][]  | A collection of matching rules to use for deciding which processes to monitor. Each config can match multiple processes, which will be tracked as a single process "group." | no

[process_names]: #process_names-block

### process_names block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name`       | `string`                         | The name to use for identifying the process group name in the metric. By default, it uses the base path of the executable. (See below for template variable details). | `{{.ExeBase}}` | no
`comm`       | `list(string)`                   | A list of strings that match the base executable name for a process, truncated to 15 characters. It is derived from reading the second field of `/proc/<pid>/stat`, stripped of parens. | | no
`exe`        | `list(string)`                   | A list of strings that match `argv[0]` for a process. If there are no slashes, only the basename of `argv[0]` needs to match. Otherwise, the name must be an exact match. For example, "postgres" may match any postgres binary, but `/usr/local/bin/postgres` will only match a postgres process with that exact path. If any of the strings match, the process will be tracked. | | no
`cmdline`    | `list(string)`                   | A list of regular expressions applied to the `argv` of the process. Each regex here must match the corresponding argv for the process to be tracked. The first element that is matched is `argv[1]`. Regex captures are added to the .Matches map for use in the name. | | no

The `name` argument can use the following template variables: 
- `{{.Comm}}`:      Basename of the original executable from /proc/\<pid\>/stat.
- `{{.ExeBase}}`:   Basename of the executable from argv[0].
- `{{.ExeFull}}`:   Fully qualified path of the executable.
- `{{.Username}}`:  Username of the effective user.
- `{{.Matches}}`:   Map containing all regex capture groups resulting from matching a process with the cmdline rule group.
- `{{.PID}}`:       PID of the process. Note that the PID is copied from the first executable found.
- `{{.StartTime}}`: The start time of the process. This is useful when combined with PID as PIDS get reused over time.

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `process_exporter` metrics.

For example, the `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metric's label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.integration.process_exporter` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.integration.process_exporter` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.integration.process_exporter` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.integration.process_exporter`:

```river
prometheus.integration.process_exporter "example" {
  track_children = false
  process_names {
    comm = ["grafana-agent"]
  }
}

// Configure a prometheus.scrape component to collect process_exporter metrics.
prometheus.scrape "demo" {
  targets    = prometheus.integration.process_exporter.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
