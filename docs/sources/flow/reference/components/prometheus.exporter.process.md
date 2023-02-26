---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.​process
---

# prometheus.exporter.process
The `prometheus.exporter.process` component embeds
[process_exporter](https://github.com/ncabatoff/process-exporter) for collecting process stats from `/proc`.

## Usage

```river
prometheus.exporter.process "LABEL" {
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`procfs_path`       | `string`                 | procfs mountpoint. | `/proc` | no
`track_children`    | `bool`                   | Whether to track a process' children. | `true` | no
`track_threads`     | `bool`                   | Report metrics for a process' individual threads.  | `true` | no
`gather_smaps`      | `bool`                   | Gather metrics from the smaps file for a process. | `true` | no
`recheck_on_scrape` | `bool`                   | Recheck process names on each scrape. | `true` | no

## Blocks
The following blocks are supported inside the definition of `prometheus.exporter.process`:

Hierarchy        | Block      | Description | Required
---------------- | ---------- | ----------- | --------
matcher          | [matcher][]  | A collection of matching rules to use for deciding which processes to monitor. | no

[matcher]: #matcher-block

### matcher block
Each `matcher` block config can match multiple processes, which will be tracked as a single process "group."

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name`       | `string`        | The name to use for identifying the process group name in the metric. | `{{.ExeBase}}` | no
`comm`       | `list(string)`  | A list of strings that match the base executable name for a process, truncated to 15 characters.  | | no
`exe`        | `list(string)`  | A list of strings that match `argv[0]` for a process. | | no
`cmdline`    | `list(string)`  | A list of regular expressions applied to the `argv` of the process. | | no

The `name` argument can use the following template variables. By default it uses the base path of the executable: 
- `{{.Comm}}`:      Basename of the original executable from /proc/\<pid\>/stat.
- `{{.ExeBase}}`:   Basename of the executable from argv[0].
- `{{.ExeFull}}`:   Fully qualified path of the executable.
- `{{.Username}}`:  Username of the effective user.
- `{{.Matches}}`:   Map containing all regex capture groups resulting from matching a process with the cmdline rule group.
- `{{.PID}}`:       PID of the process. Note that the PID is copied from the first executable found.
- `{{.StartTime}}`: The start time of the process. This is useful when combined with PID as PIDS get reused over time.

The value that is used for matching `comm` list elements is derived from reading the second field of `/proc/<pid>/stat`, stripped of parens.

For values in `exe`, if there are no slashes, only the basename of `argv[0]` needs to match. Otherwise, the name must be an exact match. For example, "postgres" may match any postgres binary, but `/usr/local/bin/postgres` will only match a postgres process with that exact path. If any of the strings match, the process will be tracked.

Each regex in `cmdline` must match the corresponding argv for the process to be tracked. The first element that is matched is `argv[1]`. Regex captures are added to the .Matches map for use in the name.

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | Targets that expose `process_exporter` metrics.

For example, the `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metric's label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.exporter.process` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.process` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.process` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.process`:

```river
prometheus.exporter.process "example" {
  track_children = false
  matcher {
    comm = ["grafana-agent"]
  }
}

// Configure a prometheus.scrape component to collect process_exporter metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.process.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
