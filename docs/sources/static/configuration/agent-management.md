---
title: Agent Management
weight: 700
---

# Agent Management - Experimental

**Agent Management is under active development. Backwards incompatible changes to its API are to be expected. Feedback is much appreciated. This is a feature that MAY NOT make it production.**

Agent Management enables centralized management of fleets of Grafana Agents.

In this mode, Grafana Agent polls and dynamically reloads its configuration from a remote API server.

Remote Configurations are composed of a base configuration and a set of snippets. Snippets are applied conditionally via label matching.

## Configuration

Agent Management can be used by passing the flag `-enable-features=agent-management`. When enabled, the file referred to `-config.file` will be loaded as an agent management configuration file.

Agent Management configuration files are YAML documents which conform the following schema:

```yaml
# Agent Management configuration.
agent_management:
  # Host of the API server to connect to.
  host: <string>

  # Protocol to use when connecting to the API server (http|https).
  protocol: <string>

  # The polling interval for fetching the configuration.
  polling_interval: <string>

  # Sets the `Authorization` header on every request with the
  # configured username and password.
  basic_auth:
    [ username: <string> ]
    [ password_file: <string> ]

  # Fields specific to remote configuration.
  remote_configuration:
    # A path to a directory where the remote configuration will be cached. The directory must be writeable.
    cache_location: <string>

    # The namespace to use.
    namespace: <string>

    # Set of self-identifying labels used for snippet selection.
    labels:
      [ <labelname>: <labelvalue> ... ]
```

## API (v2)

Grafana Agents with Agent Management enabled continuously poll the API server for an up-to-date configuration. The API server is expected to implement a `GET /agent-management/api/agent/v2/namespace/:namespace/remote_config` HTTP endpoint returning a successful response with the following body format:

```yaml
# The base configuration for the Agent.
base_config: |
  <grafana_agent_config>
# A set of snippets to be conditionally merged into the base configuration.
snippets:
  [ <snippet_name>: <snippet_content> ... ]
```

### grafana_agent_config

This is a standard Grafana Agent [static mode configuration](https://grafana.com/docs/agent/latest/static/configuration/). Typically used to configure the server, remote_writes, and other global configuration.

### snippet_content

The snippet content is a YAML document which conforms to the following schema:

```yaml
# Config provides the actual snippet configuration.
config: |
  [metrics_scrape_configs]:
  - [<scrape_config> ... ]
  [logs_scrape_configs]:
  - [<promtail.scrape_config> ... ]
  [integration_configs]:
    [<integrations_config> ... ]
# Selector is a set of labels used to decide which snippets to apply to the final configuration.
selector:
  [ <labelname>: <labelvalue> ... ]
```

> **Note:** More information on the following types can be found in their respective documentation pages:
>
> * [`scrape_config`](https://prometheus.io/docs/prometheus/2.42/configuration/configuration/#scrape_config)
> * [`promtail.scrape_config`](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#scrape_configs)
> * [`integrations_config`](https://grafana.com/docs/agent/latest/static/configuration/integrations)


### Example response body

```yaml
base_config: |
  server:
    log_level: info
  metrics:
    global:
      remote_write:
        - basic_auth:
            password_file: key.txt
            username: 123
          url: https://myserver.com/api/prom/push
  logs:
    positions_directory: /var/lib/grafana-agent
    global:
      clients:
        - basic_auth:
            password_file: key.txt
            username: 456
          url: https://myserver.com/loki/api/v1/push
snippets:
  snip1:
    config: |
      metrics_scrape_configs:
      - job_name: 'prometheus'
        scrape_interval: 15s
        static_configs:
        - targets: ['localhost:9090']
      logs_scrape_configs:
      - job_name: 'loki'
        static_configs:
        - targets: ['localhost:3100']
      integration_configs:
        node_exporter:
          enabled: true
    selector:
      os: linux
      app: app1
```

Note: This example uses YAML. However, both JSON and YAML responses are supported. The choice of serialization format is left to the server-side implementation. Grafana Agent does not send nor reads `Content-Type` or `Accept` headers.
