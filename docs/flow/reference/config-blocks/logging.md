---
aliases:
- /docs/agent/latest/flow/reference/config-blocks/logging
title: logging
weight: 100
---

# `logging` block

`logging` is a configuration block used to customize how Grafana Agent produces
log messages.

> **NOTE**: Configuration blocks are not components, so expressions which
> reference the exports of components may not be used.

## Example

```river
logging {
  level  = "error"
  format = "logfmt"
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`level` | `string` | Level at which log lines should be written | `"info"` | no
`format` | `string` | Format to use for writing log lines | `"logfmt"` | no

### Log level

The following log levels are recognized:

* `error`: Only write logs at the "error" level.
* `warn`: Only write logs at the "warn" level or above.
* `info`: Only write logs at "info" level or above.
* `debug`: Write all logs, including "debug" level logs.

### Log format

The following log line formats are recognized:

* `logfmt`: Write logs as [logfmt][] lines.
* `json`: Write logs as JSON objects.

[logfmt]: https://brandur.org/logfmt

## Log location

Grafana Agent writes all logs to `stderr`.

When running Grafana Agent as a systemd service, logs will also be sent to
`journald`.

When running Grafana Agent as a Windows service, logs are also written as event
logs and can be viewed through the Event Viewer.

In other cases, users must redirect `stderr` of the Grafana Agent process to a
file for logs to persist on disk.
