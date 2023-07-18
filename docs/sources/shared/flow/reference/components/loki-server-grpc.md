---
aliases:
- /docs/agent/shared/flow/reference/components/loki-server-grpc/
canonical: https://grafana.com/docs/grafana/agent/latest/shared/flow/reference/components/loki-server-grpc/
headless: true
---

The `grpc` block configures the gRPC server.

The following arguments can be used to configure the `grpc` block. Any omitted
fields take their default values.

 Name                            | Type       | Description                                                                                                          | Default      | Required 
---------------------------------|------------|----------------------------------------------------------------------------------------------------------------------|--------------|----------
 `listen_address`                | `string`   | Network address on which the server will listen for new connections. Defaults to accepting all incoming connections. | `""`         | no       
 `listen_port`                   | `int`      | Port number on which the server will listen for new connections. Defaults to a random free port being assigned.      | `0`          | no       
 `conn_limit`                    | `int`      | Maximum number of simultaneous http connections. Defaults to no limit.                                               | `0`          | no       
 `max_connection_age`            | `duration` | The duration for the maximum amount of time a connection may exist before it will be closed.                         | `"infinity"` | no       
 `max_connection_age_grace`      | `duration` | An additive period after `max_connection_age` after which the connection will be forcibly closed.                    | `"infinity"` | no       
 `max_connection_idle`           | `duration` | The duration after which an idle connection should be closed.                                                        | `"infinity"` | no       
 `server_max_recv_msg_size`      | `int`      | Limit on the size of a gRPC message this server can receive (bytes).                                                 | `4MB`        | no       
 `server_max_send_msg_size`      | `int`      | Limit on the size of a gRPC message this server can send (bytes).                                                    | `4MB`        | no       
 `server_max_concurrent_streams` | `int`      | Limit on the number of concurrent streams for gRPC calls (0 = unlimited).                                            | `100`        | no       
