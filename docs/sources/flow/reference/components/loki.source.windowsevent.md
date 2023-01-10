---
aliases:
- /docs/agent/latest/flow/reference/components/loki.source.windowsevent
title: loki.source.windowsevent
---

# loki.source.windowsevent

`loki.source.windowsevent` reads events from Windows Event Logs and forwards them to other
`loki.*` components.

Multiple `loki.source.windowsevent` components can be specified by giving them
different labels.

## Usage

```river
loki.source.windowsevent "LABEL" {
  eventlog_name = EVENTLOG_NAME
  forward_to    = RECEIVER_LIST
}
```

## Arguments
The component starts a new reader and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.windowsevent` supports the following arguments:

Name         | Type                 | Description                                                                    | Default                    | Required
------------ |----------------------|--------------------------------------------------------------------------------|----------------------------| --------
`locale`    | `number`             | Locale ID for event rendering.                                                 | 0 (Windows default locale) | no
`eventlog_name`    | `string`             | Event log to read from.                                                        |                            | see below
`xpath_query`    | `string`             | Event log to read from.                                                        | *                          | see below
`bookmark_path`    | `string`             | Used to keep position in event log.                                            | DATA_PATH/bookmark.xml     | no
`poll_interval`    | `time.duration`      | How often to poll the event log.                                               | 3s                         | no
`exclude_event_data`    | `bool`               | Exclude event data.                                                            | false                      | no
`exclude_user_data`    | `bool`               | Exclude user data.                                                             | false                      | no
`user_incoming_timestamp`    | `bool`               | When false will assign the current timestamp to the log when it was processed. | false                      | no
`forward_to` | `list(LogsReceiver)` | List of receivers to send log entries to.                                      |                            | yes


> **NOTE**: eventlog_name is required if xpath_query does not specify the event log.
> The xpath_query can be defined in [short or xml form](https://docs.microsoft.com/en-us/windows/win32/wes/consuming-events).


## Component health

`loki.source.windowsevent` is only reported as unhealthy if given an invalid
configuration.

## Example

This example collects log entries from the Event Log specified in eventlog_name and 
forwards them to a `loki.write` component so they are can be written to Loki.

```river
loki.source.windowsevent "application"  {
    eventlog_name = "Application"
    forward_to = [loki.write.endpoint.receiver]
}

loki.write "endpoint" {
    endpoint {
        url ="loki:3100/api/v1/push"
    }  
}
```
