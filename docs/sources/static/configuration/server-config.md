---
aliases:
- ../../configuration/server-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/server-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/server-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/server-config/
description: Learn about server_config
title: server_config
weight: 100
---

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

# TLS configuration for the HTTP server. Required when the
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

# Windows certificate filter allows selecting client CA and server certificate from the Windows Certificate store
[windows_certificate_filter: <windows_certificate_filter_config>]
```

## windows_certificate_filter_config

The `windows_certificate_filter_config` configures the use of the Windows Certificate store. Setting cert_file, key_file, and client_ca_file are invalid settings when using the windows_certificate_filter.

```yaml
# Client configuration, optional. If nothing specific will use the default client ca root
[client: <windows_client_config>]

# Name of the store to look for the Client Certificate ex My, CA
server: <windows_server_config>
```


### windows_client_config

```yaml
# Array of issuer common names to check against
issuer_common_names:
  [- <string> ... ]

# Regular expression to match Subject name
[subject_regex: <string>]

# Client Template ID to match in ASN1 format ex "1.2.3"
[template_id: <string>]
```

### windows_server_config

```yaml
# Name of the system store to look for the Server Certificate ex LocalMachine, CurrentUser
system_store: <string>

# Name of the store to look for the Server Certificate ex My, CA
store: <string>

# Array of issuer common names to check against
issuer_common_names:
[- <string> ... ]


# Server Template ID to match in ASN1 format ex "1.2.3"
[template_id: <string>]

# How often to refresh the server certificate ex 5m, 1h
[refresh_interval: <duration>]
```
