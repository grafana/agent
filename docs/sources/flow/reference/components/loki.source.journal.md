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

Name         | Type   | Description                                                                                                                                                                                                                                | Default | Required
------------ |--------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------| --------
`format_as_json`    | `bool` | When true, log messages from the journal are passed through the pipeline as a JSON message with all of the journal entries' original  fields. When false, the log message is the text content of the MESSAGE field from the journal entry. | `false` | no
`max_age`    | `duration` | The oldest relative time from process start that will be read                                                                                                                                                                              | `"7h"` | no
`path` | `string` | Path to a directory to read entries from. Defaults to system paths (/var/log/journal and /run/log/journal) when empty.                                                                                                                     | `""` | no
`matches` | `string` | Journal matches to filter. Character (+) is not supported, only logical AND matches will be added. | `""` | no
`forward_to` | `list(LogsReceiver)` | List of receivers to send log entries to.                                                                                                                                                                                                  | | yes

> **NOTE**:  A `job` label is added with the full name of the component `loki.source.journal.LABEL`. 

## Blocks

The following blocks are supported inside the definition of `loki.source.journal`:

Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
relabel_rules | [relabel_rules][] | Relabeling rules to apply to received log entries. | no

[relabel_rules]: #relabel_rules

### relabel_rules block

{{< docs/shared lookup="flow/reference/components/rule-block.md" source="agent" >}}

Incoming messages have labels from the journal following the patten `__journal_FIELDNAME`

These labels are stripped unless a rule is created to retain the labels. An example rule is 
below.

```river
rule {
		action      = "labelmap"
		regex       = "__journal_(.*)"
		replacement = "journal_${1}"
	}
```


## Component health

`loki.source.journal` is only reported as unhealthy if given an invalid
configuration.

## Debug Metrics

* `agent_loki_source_journal_target_parsing_errors_total` (counter): Total number of parsing errors while reading journal messages.
* `agent_loki_source_journal_target_lines_total` (counter): Total number of successful journal lines read.

## Example

```river
loki.source.journal "read"  {
    forward_to = [loki.write.endpoint.receiver]
}

loki.write "endpoint" {
    endpoint {
        url ="loki:3100/api/v1/push"
    }  
}
```
