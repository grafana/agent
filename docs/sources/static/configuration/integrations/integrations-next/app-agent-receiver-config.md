---
aliases:
- ../../../../configuration/integrations/integrations-next/app-agent-receiver-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/integrations-next/app-agent-receiver-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/integrations-next/app-agent-receiver-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/integrations-next/app-agent-receiver-config/
description: Learn about app_agent_receiver_config next
title: app_agent_receiver_config next
---

# app_agent_receiver_config next

The `app_agent_receiver_config` block configures the `app_agent_receiver`
integration. This integration exposes a http endpoint that can receive telemetry
from the [Grafana Faro Web SDK](https://github.com/grafana/faro-web-sdk)
and forward it to logs, traces or metrics backends.

These are the options you have for configuring the app_agent_receiver integration.

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

  # Traces instance to send traces to. This assumes that you have a traces config with such instance defined
  [traces_instance: <string> | default = ""]

  # Logs instance to send logs and exceptions to. This assumes that you have a logs
  # config with the instance defined
  [logs_instance: <string> | default = ""]

  # Server config refers to the HTTP endpoint that the integration will be exposing
  # to receive data from.
  server:
    [host: <string> | default = "127.0.0.1"]
    [port: <number> | default = 12347]

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

    # If configured, incoming requests will be required to specify this key in "x-api-key" header
    [api_key: <string>]

    # Max allowed payload size in bytes for the JSON payload. Interanlly the
    # Content-Length header is used to make this check
    [max_allowed_payload_size: <number> | default = 0]

  # Labels to set for the log entry.
  # If value is specified, it will be used.
  # If value is empty and key exists in data, it's value will be used from data
  logs_labels:
    [- <key>: <string>]

  # Timeout duration when sending an entry to Loki, milliseconds
  [logs_send_timeout: <duration> | default = 2s]

  # Sourcemap configuration for enabling stack trace transformation to original source locations
  [sourcemaps: <sourcemap_config>]
```

## sourcemap_config

```yaml
# Whether agent should attempt to download compiled sources and source maps
[download: <boolean> | default = false]

# List of HTTP origins to download sourcemaps for
[download_origins: []<string> | default = ["*"]]

# Timeout for downloading compiled sources and sourcemaps
[download_timeout: <duration> | default = "1s"]

# Sourcemap locations on filesystem. Takes precedence over downloading if both methods are enabled
filesystem:
  [- <sourcemap_file_location>]
```

## sourcemap_file_location

```yaml
# Source URL prefix. If a minified source URL matches this prefix,
# a filepath is constructed by removing the prefix, prepending path below and appending ".map".
#
# Example:
#
# minified_path_prefix = "https://my-app.dev/static/"
# path = "/var/app/static/"
#
# Then given source url "https://my-app.dev/static/foo.js"
# it will look for sourcemap at "/var/app/static/foo.js.map"

minified_path_prefix: <string>

# Directory on file system that contains source maps.
# See above for more detailed explanation.
# It is parsed as a Go template. You can use "{{.Release }}" which will be replaced with
# app.release meta property.
path: <string>
```
