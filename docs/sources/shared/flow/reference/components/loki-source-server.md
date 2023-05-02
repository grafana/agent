---
title: loki.server
---

# loki.server

`loki.server` is a common block used in all `loki.` components that expose a network server, such as
[loki.source.gcplog][] and [loki.source.heroku][]. The
network server supports both `HTTP` and `gRPC` protocols.

[loki.source.gcplog]: {{< relref "./loki.source.gcplog.md" >}}
[loki.source.heroku]: {{< relref "./loki.source.heroku.md" >}}

## Usage

```river
server {
    http {
        listen_port = 8080
        listen_adress = "0.0.0.0"
    }
    grpc {
        listen_port = 8081
    }
}
```

## Arguments

`loki.server` supports the following arguments:

 Name                        | Type           | Description                               | Default | Required
-----------------------------|----------------|-------------------------------------------|---------|----------
 `graceful_shutdown_timeout` | `duration` | Timeout for graceful shutdowns. | "30s"    | no

## Blocks

The following blocks are supported inside the definition of
`loki.source.gcplog`:

 Hierarchy | Name     | Description                 | Required
-----------|----------|-----------------------------|----------
 http      | [http][] | Configures the HTTP server. | no
 grpc      | [grpc][] | Configures the gRPC server. | no

The `pull` and `push` inner blocks are mutually exclusive; a component must
contain exactly one of the two in its definition.

[http]: #http-block

[grpc]: #grpc-block

### http block

The `http` configures the HTTP server.

The following arguments can be used to configure the `http` block. Any omitted
fields take their default values.

 Name                   | Type       | Description                                                                                                          | Default  | Required
------------------------|------------|----------------------------------------------------------------------------------------------------------------------|----------|----------
 `listen_address`       | `string`   | Network address on which the server will listen for new connections. Defaults to accepting all incoming connections. | `""`     | no
 `listen_port`          | `int`      | Port number on which the server will listen for new connections.                                                     | `8080`   | no
 `conn_limit`           | `int`      | Maximum number of simultaneous http connections. Defaults to no limit.                                               | `0`      | no
 `server_read_timeout`  | `duration` | Read timeout for HTTP server.                                                                                        | `"30s"`  | no
 `server_write_timeout` | `duration` | Write timeout for HTTP server.                                                                                       | `"30s"`  | no
 `server_idle_timeout`  | `duration` | Idle timeout for HTTP server.                                                                                        | `"120s"` | no

### grpc block

The `grpc` configures the gRPC server.

The following arguments can be used to configure the `grpc` block. Any omitted
fields take their default values.

 Name                            | Type       | Description                                                                                                          | Default      | Required
---------------------------------|------------|----------------------------------------------------------------------------------------------------------------------|--------------|----------
 `listen_address`                | `string`   | Network address on which the server will listen for new connections. Defaults to accepting all incoming connections. | `""`         | no
 `listen_port`                   | `int`      | Port number on which the server will listen for new connections.                                                     | `8081`       | no
 `conn_limit`                    | `int`      | Maximum number of simultaneous http connections. Defaults to no limit.                                               | `0`          | no
 `max_connection_age`            | `duration` | The duration for the maximum amount of time a connection may exist before it will be closed.                         | `"infinity"` | no
 `max_connection_age_grace`      | `duration` | An additive period after `max_connection_age` after which the connection will be forcibly closed.                    | `"infinity"` | no
 `max_connection_idle`           | `duration` | The duration after which an idle connection should be closed.                                                        | `"infinity"` | no
 `server_max_recv_msg_size`      | `int`      | Limit on the size of a gRPC message this server can receive (bytes).                                                 | `4MB`        | no
 `server_max_send_msg_size`      | `int`      | Limit on the size of a gRPC message this server can send (bytes).                                                    | `4MB`        | no
 `server_max_concurrent_streams` | `int`      | Limit on the number of concurrent streams for gRPC calls (0 = unlimited).                                            | `100`        | no
