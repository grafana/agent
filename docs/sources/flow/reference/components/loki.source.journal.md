---
title: loki.source.journal
---

# loki.source.journal

`loki.source.journal` reads from the systemd journal and forwards them to other
`loki.*` components.

Multiple `loki.source.journal` components can be specified by giving them
different labels.

## Usage

```river
loki.source.journal "LABEL" {
  forward_to    = RECEIVER_LIST
}
```

## Arguments
The component starts a new journal reader and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.journal` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`format_as_json` | `bool` | Whether to forward the original journal entry as JSON. | `false` | no
`max_age` | `duration` | The oldest relative time from process start that will be read. | `"7h"` | no
`path` | `string` | Path to a directory to read entries from. | `""` | no
`matches` | `string` | Journal matches to filter. The `+` character is not supported, only logical AND matches will be added. | `""` | no
`forward_to` | `list(LogsReceiver)` | List of receivers to send log entries to. | | yes
`relabel_rules` | `RelabelRules` | Relabeling rules to apply on log entries. | `{}` | no

> **NOTE**:  A `job` label is added with the full name of the component `loki.source.journal.LABEL`.

When the `format_as_json` argument is true, log messages are passed through as
JSON with all of the original fields from the journal entry. Otherwise, the log
message is taken from the content of the `MESSAGE` field from the journal
entry.

When the `path` argument is empty, `/var/log/journal` and `/run/log/journal`
will be used for discovering journal entries.

The `relabel_rules` field can make use of the `rules` export value from a
`loki.relabel` component to apply one or more relabeling rules to log entries
before they're forwarded to the list of receivers in `forward_to`.

All messages read from the journal include internal labels following the
pattern of `__journal_FIELDNAME` and will be dropped before sending to the list
of receivers specified in `forward_to`. To keep these labels, use the
`relabel_rules` argument and relabel them to not be prefixed with `__`.

> **NOTE**: many field names from journald start with an `_`, such as
> `_systemd_unit`. The final internal label name would be
> `__journal__systemd_unit`, with _two_ underscores between `__journal` and
> `systemd_unit`.

## Component health

`loki.source.journal` is only reported as unhealthy if given an invalid
configuration.

## Debug Metrics

* `agent_loki_source_journal_target_parsing_errors_total` (counter): Total number of parsing errors while reading journal messages.
* `agent_loki_source_journal_target_lines_total` (counter): Total number of successful journal lines read.

## Example

```river
loki.relabel "journal" {
  rule {
    source_labels = ["__journal__systemd_unit"]
    target_label  = "unit"
  }
}

loki.source.journal "read"  {
  forward_to    = [loki.write.endpoint.receiver]
  relabel_rules = loki.relabel.journal.rules
}

loki.write "endpoint" {
  endpoint {
    url ="loki:3100/api/v1/push"
  }
}
```
