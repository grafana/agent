---
aliases:
- ../../../configuration/integrations/process-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/process-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/process-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/process-exporter-config/
description: Learn about process_exporter_config
title: process_exporter_config
---

# process_exporter_config

The `process_exporter_config` block configures the `process_exporter` integration,
which is an embedded version of
[`process-exporter`](https://github.com/ncabatoff/process-exporter)
and allows for collection metrics based on the /proc filesystem on Linux
systems. Note that on non-Linux systems, enabling this exporter is a no-op.

Note that if running the Agent in a container, you will need to bind mount
folders from the host system so the integration can monitor them:

```
docker run \
  -v "/proc:/proc:ro" \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:{{< param "AGENT_RELEASE" >}} \
  --config.file=/etc/agent-config/agent.yaml
```

Replace `/path/to/config.yaml` with the appropriate path on your host system
where an Agent config file can be found.

For running on Kubernetes, ensure to set the equivalent mounts and capabilities
there as well:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: agent
spec:
  containers:
  - image: grafana/agent:{{< param "AGENT_RELEASE" >}}
    name: agent
    args:
    - --config.file=/etc/agent-config/agent.yaml
    volumeMounts:
    - name: procfs
      mountPath: /proc
      readOnly: true
  volumes:
  - name: procfs
    hostPath:
      path: /proc
```

The manifest and Tanka configs provided by this repository do not have the
mounts or capabilities required for running this integration.

An example config for `process_exporter_config` that tracks all processes is the
following:

```
enabled: true
process_names:
- name: "{{.Comm}}"
  cmdline:
  - '.+'
```

Full reference of options:

```yaml
  # Enables the process_exporter integration, allowing the Agent to automatically
  # collect system metrics from the host UNIX system.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the agent hostname
  # and HTTP listen port, delimited by a colon.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the process_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/process_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # procfs mountpoint.
  [procfs_path: <string> | default = "/proc"]

  # If a proc is tracked, track with it any children that aren't a part of their
  # own group.
  [track_children: <boolean> | default = true]

  # Report on per-threadname metrics as well.
  [track_threads: <boolean> | default = true]

  # Gather metrics from smaps file, which contains proportional resident memory
  # size.
  [gather_smaps: <boolean> | default = true]

  # Recheck process names on each scrape.
  [recheck_on_scrape: <boolean> | default = false]

  # A collection of matching rules to use for deciding which processes to
  # monitor. Each config can match multiple processes to be tracked as a single
  # process "group."
  process_names:
    [- <process_matcher_config>]
```

## process_matcher_config

```yaml
# The name to use for identifying the process group name in the metric. By
# default, it uses the base path of the executable.
#
# The following template variables are available:
#
# - {{.Comm}}:      Basename of the original executable from /proc/<pid>/stat
# - {{.ExeBase}}:   Basename of the executable from argv[0]
# - {{.ExeFull}}:   Fully qualified path of the executable
# - {{.Username}}:  Username of the effective user
# - {{.Matches}}:   Map containing all regex capture groups resulting from
#                   matching a process with the cmdline rule group.
# - {{.PID}}:       PID of the process. Note that the PID is copied from the
#                   first executable found.
# - {{.StartTime}}: The start time of the process. This is useful when combined
#                   with PID as PIDS get reused over time.
# - `{{.Cgroups}}`: The cgroups, if supported, of the process (`/proc/self/cgroup`). This is particularly useful for identifying to which container a process belongs.
#
# **NOTE**: Using `PID` or `StartTime` is discouraged, as it is almost never what you want, and is likely to result in high cardinality metrics.


[name: <string> | default = "{{.ExeBase}}"]

# A list of strings that match the base executable name for a process, truncated
# at 15 characters. It is derived from reading the second field of
# /proc/<pid>/stat minus the parens.
#
# If any of the strings match, the process will be tracked.
comm:
  [- <string>]

# A list of strings that match argv[0] for a process. If there are no slashes,
# only the basename of argv[0] needs to match. Otherwise the name must be an
# exact match. For example, "postgres" may match any postgres binary but
# "/usr/local/bin/postgres" can only match a postgres at that path exactly.
#
# If any of the strings match, the process will be tracked.
exe:
  [- <string>]

# A list of regular expressions applied to the argv of the process. Each
# regex here must match the corresponding argv for the process to be tracked.
# The first element that is matched is argv[1].
#
# Regex Captures are added to the .Matches map for use in the name.
cmdline:
  [- <string>]
```
