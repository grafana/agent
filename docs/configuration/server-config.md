+++
title = "server_config"
weight = 100
+++

# server_config

The `server_config` block configures the Agent's behavior as an HTTP server,
gRPC server, and the log level for the whole process.

The Agent exposes an HTTP server for scraping its own metrics and gRPC for the
scraping service mode.

```yaml
# HTTP server listen host. Used for Agent metrics, integrations, and the Agent
# API.
[http_listen_address: <string> | default = "0.0.0.0"]

# HTTP server listen port
[http_listen_port: <int> | default = 80]

# gRPC server listen host. Used for clustering, but runs even when
# clustering is disabled.
[grpc_listen_address: <string> | default = "0.0.0.0"]

# gRPC server listen port. Used for clustering, but runs even when
# clustering is disabled.
[grpc_listen_port: <int> | default = 9095]

# Register instrumentation handlers (/metrics, etc.)
[register_instrumentation: <boolean> | default = true]

# Timeout for graceful shutdowns
[graceful_shutdown_timeout: <duration> | default = 30s]

# Read timeout for HTTP server
[http_server_read_timeout: <duration> | default = 30s]

# Write timeout for HTTP server
[http_server_write_timeout: <duration> | default = 30s]

# Idle timeout for HTTP server
[http_server_idle_timeout: <duration> | default = 120s]

# Max gRPC message size that can be received. Unused.
[grpc_server_max_recv_msg_size: <int> | default = 4194304]

# Max gRPC message size that can be sent. Unused.
[grpc_server_max_send_msg_size: <int> | default = 4194304]

# Limit on the number of concurrent streams for gRPC calls (0 = unlimited).
# Unused.
[grpc_server_max_concurrent_streams: <int> | default = 100]

# Log only messages with the given severity or above. Supported values [debug,
# info, warn, error]. This level affects logging for the whole application, not
# just the Agent's HTTP/gRPC server.
[log_level: <string> | default = "info"]

# Base path to server all API routes from (e.g., /v1/). Unused.
[http_path_prefix: <string>]

# Configuration for HTTPS serving and scraping of metrics
[http_tls_config: <server_tls_config>]
```

## server_tls_config

The `http_tls_config` block configures the server to run with TLS. When set, `integrations.http_tls_config` must
also be provided. Acceptable values for  `client_auth_type` are found in
[Go's `tls` package](https://golang.org/pkg/crypto/tls/#ClientAuthType).

```yaml
# File path to the server certificate
[cert_file: <string>]

# File path to the server key
[key_file: <string>]

# Tells the server what is acceptable from the client, this drives the options in client_tls_config
[client_auth_type: <string>]

# File path to the signing CA certificate, needed if CA is not trusted
[client_ca_file: <string>]
```
