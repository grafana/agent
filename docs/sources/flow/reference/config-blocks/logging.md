---
title: logging
---

# logging block

`logging` is an optional configuration block used to customize how Grafana
Agent produces log messages. `logging` is specified without a label and can
only be provided once per configuration file.

## Example

```river
logging {
  level  = "info"
  format = "logfmt"
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`level` | `string` | Level at which log lines should be written | `"info"` | no
`format` | `string` | Format to use for writing log lines | `"logfmt"` | no
`write_to` | `list(LogsReceiver)` | List of receivers to send log entries to |  | no

### Log level

The following strings are recognized as valid log levels:

* `"error"`: Only write logs at the _error_ level.
* `"warn"`: Only write logs at the _warn_ level or above.
* `"info"`: Only write logs at _info_ level or above.
* `"debug"`: Write all logs, including _debug_ level logs.

### Log format

The following strings are recognized as valid log line formats:

* `"logfmt"`: Write logs as [logfmt][] lines.
* `"json"`: Write logs as JSON objects.

[logfmt]: https://brandur.org/logfmt

### Log receivers

The `write_to` argument allows the Agent to tee its log entries to one or more
`loki.*` component log receivers in addition to the default [location][].
This, for example can be the export of a `loki.write` component to ship log
entries directly to Loki, or a `loki.relabel` component to add a certain label
first.

[location]: #log-location

## Log location

Grafana Agent writes all logs to `stderr`.

When running Grafana Agent as a systemd service, view logs written to `stderr`
through `journald`.

When running Grafana Agent as a container, view logs written to `stderr`
through `docker logs` or `kubectl logs`, depending on whether Docker or
Kubernetes was used for deploying the agent.

When running Grafana Agent as a Windows service, logs are instead written as
event logs; view logs through Event Viewer.

In other cases, redirect `stderr` of the Grafana Agent process to a file for
logs to persist on disk.
