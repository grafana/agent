+++
title = "app_o11y_config"
+++

# app_o11y_config

The `app_o11y_config` block configures the `app-o11y-exporter`
integration, which is a data exporter for logs/metrics/traces/exceptions
received by the respectful JavaScript agent

Full reference of options:

```yaml
  # Enables the app_o11y_config integration, allowing the Agent to automatically
  # collect metrics,logs and exceptions from the grafana frontend agent
  [enabled: <boolean> | default = false]

  # Domains in which the agent is sending data from
  cors_allowed_origins:
    [- <string>]

  # Configure rate limiting. The HTTP server of the App observability implements
  # a token bucket rate limitng algorithm in which we can configure the maximum RPS
  # as well as the burstiness (peaks of RPS)
  rate_limiting:
    [enabled: <boolean> | default = false]
    [rps: <number> | default = 100]
    [burstiness: <number> | default = 50]

  # Max allowed payload size in bytes for the JSON payload. Interanlly the
  # Content-Length header is used to make this check
  [max_allowed_payload_size: <number> | default = 0]

  # Configure source map support. The app observability integration can do a basic
  # deobfuscation of JavaScript stacktraces using a user defined map file. The map_uri
  # can be either a filesystem path or a URL to download the file from
  source_map:
    [enabled: <boolean> | default = false]
    [map_uri: <string> | default = ""]

  # Loki instance to send logs and exceptions to. This assumes that you have a logs
  # config with the instance defined
  [logs_instance: <string> | default = "default"]

  # Server config refres to the HTTP endpoint that the integration will be exposing
  # to receive data from.
  server:
    [host: <string> | default = "0.0.0.0"]
    [port: <number> | port = 8080]

  # User defined prometheus metrics to be scraped. The sending end of the agent can
  # update this metrics using the specified payload. Since these are performance metrics
  # a Summary type metric is used internally for each metric defined
  custom_measurements:
    [name: <string>]
    [description: <string>]

```
