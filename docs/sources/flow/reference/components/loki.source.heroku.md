---
title: loki.source.heroku
---

# loki.source.heroku

`loki.source.heroku` listens for Heroku messages over TCP connections
and forwards them to other `loki.*` components.

The component starts a new heroku listener for the given `listener`
block and fans out incoming entries to the list of receivers in `forward_to`.

Before using `loki.source.heroku`, Heroku should be configured with the URL where the Agent will be listening. Follow the steps in [Heroku HTTPS Drain docs](https://devcenter.heroku.com/articles/log-drains#https-drains) for using the Heroku CLI with a command like the following:

```shell
heroku drains:add [http|https]://HOSTNAME:PORT/heroku/api/v1/drain -a HEROKU_APP_NAME
```

Multiple `loki.source.heroku` components can be specified by giving them
different labels.

## Usage

```river
loki.source.heroku "LABEL" {
    listener {
        address = "LISTEN_ADDRESS"
        port    = PORT
    }
    forward_to = RECEIVER_LIST
}
```

## Arguments

`loki.source.heroku` supports the following arguments:

Name                     | Type                   | Description          | Default | Required
------------------------ | ---------------------- | -------------------- | ------- | --------
`use_incoming_timestamp` | `bool`                 | Whether or not to use the timestamp received from Heroku. | `false` | no
`labels`                 | `map(string)`          | The labels to associate with each received Heroku record. | `{}`    | no
`forward_to`             | `list(LogsReceiver)`   | List of receivers to send log entries to.                 |         | yes
`relabel_rules`          | `RelabelRules`         | Relabeling rules to apply on log entries.                 | `{}`    | no

The `relabel_rules` field can make use of the `rules` export value from a
`loki.relabel` component to apply one or more relabeling rules to log entries
before they're forwarded to the list of receivers in `forward_to`.

## Blocks

The following blocks are supported inside the definition of `loki.source.heroku`:

<<<<<<< HEAD
Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
listener | [listener] | Configures a listener for Heroku messages. | yes
=======
 Hierarchy | Name     | Description                                        | Required 
-----------|----------|----------------------------------------------------|----------
 `http`    | [http][] | Configures the HTTP server that receives requests. | no       
 `grpc`    | [grpc][] | Configures the gRPC server that receives requests. | no       
>>>>>>> 0308a3270... Fix panic when config not provided for `loki.source.(heroku|gcplog)` (#3776)

[listener]: #listener-block

### listener block

The `listener` block defines the listen address and port where the listener
expects Heroku messages to be sent to.

Name                     | Type          | Description | Default | Required
------------------------ | ------------- | ----------- | ------- | --------
`address`                | `string`      | The `<host>` address to listen to for heroku messages. | `0.0.0.0` | no
`port`                   | `int`         | The `<port>` to listen to for heroku messages. | | yes

## Labels

The `labels` map is applied to every message that the component reads.

The following internal labels all prefixed with `__` are available but will be discarded if not relabeled:
- `__heroku_drain_host`
- `__heroku_drain_app`
- `__heroku_drain_proc`
- `__heroku_drain_log_id`

All url query params will be translated to `__heroku_drain_param_<name>`

If the `X-Scope-OrgID` header is set it will be translated to `__tenant_id__`

## Exported fields

`loki.source.heroku` does not export any fields.

## Component health

`loki.source.heroku` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.heroku` exposes some debug information per Heroku listener:
* Whether the listener is currently running.
* The listen address.

## Debug metrics
* `loki_source_heroku_drain_entries_total` (counter): Number of successful entries received by the Heroku target.
* `loki_source_heroku_drain_parsing_errors_total` (counter): Number of parsing errors while receiving Heroku messages.

## Example

This example listens for Heroku messages over TCP in the specified port and forwards them to a `loki.write` component using the Heroku timestamp.

```river
loki.source.heroku "local" {
    listener {
        address = "0.0.0.0"
        port    = 8080
    }
    use_incoming_timestamp = true
    labels                 = {component = "loki.source.heroku"}
    forward_to             = [loki.write.local.receiver]
}

loki.write "local" {
    endpoint {
        url = "loki:3100/api/v1/push"
    }
}
```

