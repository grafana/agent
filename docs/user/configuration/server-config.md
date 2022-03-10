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
# Log only messages with the given severity or above. Supported values [debug,
# info, warn, error]. This level affects logging for all Agent-level logs, not
# just the HTTP and gRPC server.
#
# Note that some integrations use their own loggers which ignore this
# setting.
[log_level: <string> | default = "info"]

# Log messages with the given format. Supported values [logfmt, json].
# This affects logging for all Agent-levle logs, not just the HTTP and gRPC
# server.
#
# Note that some integrations use their own loggers which ignore this
# setting.
[log_format: <string> | default = "logfmt"]

# TLS configuration for the HTTP server. Reuqired when the
# -server.http.tls-enabled flag is provided, ignored otherwise.
[http_tls_config: <server_tls_config>]

# TLS configuration for the gRPC server. Required when the
# -server.grpc.tls-enabled flag is provided, ignored otherwise.
[grpc_tls_config: <server_tls_config>]
```

## server_tls_config

The `server_tls_config` configures TLS.

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
