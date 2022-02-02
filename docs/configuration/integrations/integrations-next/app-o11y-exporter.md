+++
title = "app_o11y_config"
+++

# app_o11y_receiver_config

The `app_o11y_receiver_config` block configures the `app-o11y-receiver`
integration. This integration exposes a http endpoint that can receive telemetry
from the [grafana javascript agent](https://github.com/grafana/grafana-javascript-agent)
and forward it to logs, traces or metrics backends.

Full reference of options:

```yaml
  autoscrape:
    # Enables autoscrape of integrations.
    [enable: <boolean> | default = true]

    # Specifies the metrics instance name to send metrics to. Instance
    # names are located at metrics.configs[].name from the top-level config.
    # The instance must exist.
    #
    # As it is common to use the name "default" for your primary instance,
    # we assume the same here.
    [metrics_instance: <string> | default = "default"]

    # Autoscrape interval and timeout. Defaults are inherited from the global
    # section of the top-level metrics config.
    [scrape_interval: <duration> | default = <metrics.global.scrape_interval>]
    [scrape_timeout: <duration> | default = <metrics.global.scrape_timeout>]

  # Integration instance name
  [instance: <string>]

  # Domains in which the agent is sending data from. For example "https://myapp.com"
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

  # Loki instance to send logs and exceptions to. This assumes that you have a logs
  # config with the instance defined
  [logs_instance: <string> | default = "default"]

  # Server config refres to the HTTP endpoint that the integration will be exposing
  # to receive data from.
  server:
    [host: <string> | default = "0.0.0.0"]
    [port: <number> | default = 8080]

  # Labels to set for the log entry. 
  # If value is specified, it will be used.
  # If value is empty and key exists in data, it's value will be used from data
  logs_labels:
    [- <key>: <string>]

  # Timeout duration when sending an entry to Loki
  [logs_send_timeout: <number> | default = 2000]


```
