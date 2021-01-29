# Migration Guide

This is a guide detailing all breaking changes that have happened in prior
releases and how to migrate to newer versions.

# v0.12.0

v0.12.0 had two breaking changes: the `tempo` and `loki` sections have been changed to require a list of `tempo`/`loki` configs rather than just one.

## Tempo Config Change

The Tempo config (`tempo` in the config file) has been changed to store
configs within a `configs` list. This allows for defining multiple Tempo
instances for collecting traces and forwarding them to different OTLP
endpoints.

To migrate, add a `configs:` array and move your existing config inside of it.
Give the element a `name: default` field.

Each config must have a unique non-empty name. `default` is recommended for users
that don't have other configs. The name of the config will be added as a
`tempo_config` label for metrics.

Example old config:

```yaml
tempo:
  receivers:
    jaeger:
      protocols:
        thrift_http:
  attributes:
    actions:
    - action: upsert
      key: env
      value: prod
  push_config:
    endpoint: otel-collector:55680
    insecure: true
    batch:
      timeout: 5s
      send_batch_size: 100
```

Example migrated config:

```yaml
tempo:
  configs:
  - name: default
    receivers:
      jaeger:
        protocols:
          thrift_http:
    attributes:
      actions:
      - action: upsert
        key: env
        value: prod
    push_config:
      endpoint: otel-collector:55680
      insecure: true
      batch:
        timeout: 5s
        send_batch_size: 100
```

## Loki Promtail Config Change

The Loki Promtail config (`loki` in the config file) has been changed to store
configs within a `configs` list. This allows for defining multiple Loki
Promtail instances for collecting logs and forwarding them to different Loki
servers.

To migrate, add a `configs:` array and move your existing config inside of it.
Give the element a `name: default` field.

Each config must have a unique non-empty name. `default` is recommended for users
that don't have other configs. The name of the config will be added as a
`loki_config` label for Loki Promtail metrics.

Example old config:

```yaml
loki:
  positions:
    filename: /tmp/positions.yaml
  clients:
    - url: http://loki:3100/loki/api/v1/push
  scrape_configs:
  - job_name: system
    static_configs:
      - targets:
        - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log
```

Example migrated config:

```yaml
loki:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    clients:
      - url: http://loki:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
        - targets:
          - localhost
          labels:
            job: varlogs
            __path__: /var/log/*log
```
