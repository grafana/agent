---
title: loki.source.gelf
---

# loki.source.gelf

`loki.source.gelf` reads [Graylog Extended Long Format (GELF) logs](https://github.com/Graylog2/graylog2-server) from a UDP listener and forwards them to other
`loki.*` components. 

Multiple `loki.source.gelf` components can be specified by giving them
different labels and ports.

## Usage

```river
loki.source.gelf "LABEL" {
  forward_to    = RECEIVER_LIST
}
```

## Arguments
The component starts a new UDP listener and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.gelf` supports the following arguments:

Name         | Type                 | Description                                                                    | Default                    | Required
------------ |----------------------|--------------------------------------------------------------------------------|----------------------------| --------
`listen_address`    | `string`             | UDP address and port to listen for Graylog messages.                    | `0.0.0.0:12201` | no
`use_incoming_timestamp`    | `bool`             | When false, assigns the current timestamp to the log when it was processed | `false`                            | no
`relabel_rules` | `RelabelRules`         | Relabeling rules to apply on log entries. | "{}" | no


> **NOTE**: GELF logs can be sent uncompressed or compressed with GZIP or ZLIB. 
> A `job` label is added with the full name of the component `loki.source.gelf.LABEL`. 

Incoming messages have the following labels available:
* `__gelf_message_level`: The GELF level as a string.
* `__gelf_message_host`: The host sending the GELF message.
* `__gelf_message_host`: The GELF level message version sent by the client.
* `__gelf_message_facility`: The GELF facility.

These labels are stripped unless relabeling is used to retain the labels with a
`loki.relabel` component and its Rules export as the `relabel_rules` argument.
An example rule is presented below.

```river
rule {
		action      = "labelmap"
		regex       = "__gelf_(.*)"
		replacement = "gelf_${1}"
	}
```


## Component health

`loki.source.gelf` is only reported as unhealthy if given an invalid
configuration.

## Debug Metrics

* `gelf_target_entries_total` (counter): Total number of successful entries sent to the GELF target.
* `gelf_target_parsing_errors_total` (counter): Total number of parsing errors while receiving GELF messages.

## Example

```river
loki.source.gelf "listen"  {
    forward_to = [loki.write.endpoint.receiver]
}

loki.write "endpoint" {
    endpoint {
        url ="loki:3100/api/v1/push"
    }  
}
```
