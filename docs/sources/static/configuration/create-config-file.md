---
aliases:
- ../../configuration/create-config-file/
- ../../set-up/create-config-file/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/create-config-file/
- /docs/grafana-cloud/send-data/agent/static/configuration/create-config-file/
canonical: https://grafana.com/docs/agent/latest/static/configuration/create-config-file/
description: Learn how to create a configuration file
title: Create a configuration file
weight: 50
---

# Create a configuration file

The Grafana Agent supports configuring multiple independent "subsystems." Each
subsystem helps you collect data for a specific type of telemetry.

- The **Metrics** subsystem allows you collect metrics to send to Prometheus.
- The **Logs** subsystem allows you to collect logs to send to Grafana Loki.
- The **Traces** subsystem allows you to collect spans to send to Grafana Tempo.
- The **Integrations** subsystem allows you to collect metrics for common
  applications, such as MySQL.

Integrations are recommended for first-time users of observability platforms,
especially newcomers to Prometheus. Users with more experience with Prometheus
or users that already have an existing Prometheus config file can configure
the Prometheus subsystem manually.

## Integrations

_Integrations_ are individual features that collect metrics for you. For
example, the `agent` integration collects metrics from that running instance of
the Grafana Agent. The `node_exporter` integration will collect metrics from the
Linux machine that the Grafana Agent is running on.

```yaml
metrics:
  wal_directory: /tmp/wal
  global:
    remote_write:
      - url: http://localhost:9009/api/prom/push

integrations:
  agent:
    enabled: true
```

In this example, we first must configure the `wal_directory` which is used to
store metrics in a Write-Ahead Log (WAL). The WAL is required and ensures that samples
will be redelivered in case of failure (e.g., network issues, machine reboot). We
also configure `remote_write`, which is where all metrics should be sent by
default.

Then, the individual `integrations` are configured. In this example, just the
`agent` integration is enabled. Finally, `prometheus_remote_write` is configured
with a location to send metrics. You will have to replace this URL with the
appropriate URL for your `remote_write` system (such as a Grafana Cloud Hosted
Prometheus instance).

When the Agent is run with this file, it will collect metrics from itself and
send those metrics to the default `remote_write` endpoint. All metrics from
integrations will have an `instance` label matching the hostname of the machine
the Grafana Agent is running on. This label helps to uniquely identify the
source of metrics if you are running multiple Grafana Agents across multiple
machines.

Full configuration options can be found in the [configuration reference][configure].

## Prometheus config/migrating from Prometheus

The Prometheus subsystem config is useful for those migrating from Prometheus
and those who want to scrape metrics from something that currently does not have
an associated integration.

To migrate from an existing Prometheus config, use this Agent config as a
template and copy and paste subsections from your existing Prometheus config
into it:

```yaml
metrics:
  global:
  # PASTE PROMETHEUS global SECTION HERE
  configs:
    - name: agent
      scrape_configs:
        # PASTE scrape_configs SECTION HERE
      remote_write:
        # PASTE remote_write SECTION HERE
```

For example, this configuration file configures the Grafana Agent to
scrape itself without using the integration:

```yaml
server:
  log_level: info

metrics:
  global:
    scrape_interval: 1m
  configs:
    - name: agent
      scrape_configs:
        - job_name: agent
          static_configs:
            - targets: ['127.0.0.1:12345']
      remote_write:
        - url: http://localhost:9009/api/prom/push
```

Like with integrations, full configuration options can be found in the
[configuration][configure].

## Loki Config/Migrating from Promtail

The Loki Config allows for collecting logs to send to a Loki API. Users that are
familiar with Promtail will notice that the Loki config for the Agent matches
their existing Promtail config with the following exceptions:

- The deprecated field `client` is not present
- The `server` field is not present

To migrate from an existing Promtail config, make sure you are using `clients`
instead of `client` and remove the `server` block if present. Then paste your
Promtail config into the Agent config file inside of a `logs` section:

```yaml
logs:
  configs:
  - name: default
    # PASTE YOUR PROMTAIL CONFIG INSIDE OF HERE
```

### Full config example

Here is an example full config file, using integrations, Prometheus, Loki, and
Tempo:

```yaml
server:
  log_level: info

metrics:
  global:
    scrape_interval: 1m
    remote_write:
      - url: http://localhost:9009/api/prom/push
  configs:
    - name: default
      scrape_configs:
        - job_name: agent
          static_configs:
            - targets: ['127.0.0.1:12345']

logs:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    scrape_configs:
      - job_name: varlogs
        static_configs:
          - targets: [localhost]
            labels:
              job: varlogs
              __path__: /var/log/*log
    clients:
      - url: http://localhost:3100/loki/api/v1/push

traces:
  configs:
  - name: default
    receivers:
      jaeger:
        protocols:
          grpc: # listens on the default jaeger grpc port: 14250
    remote_write:
      - endpoint: localhost:55680
        insecure: true  # only add this if TLS is not required
    batch:
      timeout: 5s
      send_batch_size: 100

integrations:
  node_exporter:
    enabled: true
```

{{% docs/reference %}}
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration"
[configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/static/configuration"
{{% /docs/reference %}}
