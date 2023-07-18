---
aliases:
- /docs/agent/shared/flow/reference/components/loki-server-http/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/loki-server-http/
headless: true
---

The `http` block configures the HTTP server.

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
