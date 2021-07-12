+++
title = "Upgrade guide"
weight = 200
+++

# Upgrade guide

This guide describes all breaking changes that have happened in prior
releases and how to migrate to newer versions.

## Unreleased

These changes will come in a future version.

### Tempo: Remote write TLS config

Tempo `remote_write` now supports configuring TLS settings in the trace
exporter's client. `insecure_skip_verify` is moved into this setting's block.

Old configurations with `insecure_skip_verify` outside `tls_config` will continue
to work until it's fully deprecated.
If both `insecure_skip_verify` and `tls_config.insecure_skip_verify` are used,
then the latter take precedence.

Example old config:

```
tempo:
  configs:
    - name: default
      remote_write:
        - endpoint: otel-collector:55680
          insecure: true
          insecure_skip_verify: true
```

Example new config:

```
tempo:
  configs:
    - name: default
      remote_write:
        - endpoint: otel-collector:55680
          insecure: true
          tls_config:
            insecure_skip_verify: true
```

## v0.15.0

### Tempo: `automatic_logging` changes

Tempo automatic logging previously assumed that the operator wanted to log
to a Loki instance. With the addition of an option to log to stdout a new
field is required to maintain the old behavior.

Example old config:

```
tempo:
  configs:
  - name: default
    automatic_logging:
      loki_name: <some loki instance>
```

Example new config:

```
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: loki
      loki_name: <some loki instance>
```

## v0.14.0

### Scraping Service security change

v0.14.0 changes the default behavior of the scraping service config management
API to reject all configuration files that read credentials from a file on disk.
This prevents malicious users from crafting an instance config file that read
arbitrary files on disk and send their contents to remote endpoints.

To revert to the old behavior, add `dangerous_allow_reading_files: true` in your
`scraping_service` config.

Example old config:

```yaml
prometheus:
  scraping_service:
    # ...
```

Example new config:

```yaml
prometheus:
  scraping_service:
    dangerous_allow_reading_files: true
    # ...
```

### SigV4 config change

v0.14.0 updates the internal Prometheus dependency to 2.26.0, which includes
native support for SigV4, but uses a slightly different configuration structure
than the Grafana Agent did.

To migrate, remove the `enabled` key from your `sigv4` configs. If `enabled` was
the only key, define sigv4 as an empty object: `sigv4: {}`.

Example old config:

```yaml
sigv4:
  enabled: true
  region: us-east-1
```

Example new config:

```yaml
sigv4:
  region: us-east-1
```

### Tempo: `push_config` deprecation

`push_config` is now deprecated in favor of a `remote_write` array which allows for sending spans to multiple endpoints.
`push_config` will be removed in a future release, and it is recommended to migrate to `remote_write` as soon as possible.

To migrate, move the batch options outside the `push_config` block.
Then, add a `remote_write` array and move the remaining of your `push_config` block inside it.

Example old config:

```yaml
tempo:
  configs:
    - name: default
      receivers:
        otlp:
          protocols:
            gpc:
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
        otlp:
          protocols:
            gpc:
      remote_write:
        - endpoint: otel-collector:55680
          insecure: true
      batch:
        timeout: 5s
        send_batch_size: 100
```

## v0.12.0

v0.12.0 had two breaking changes: the `tempo` and `loki` sections have been changed to require a list of `tempo`/`loki` configs rather than just one.

### Tempo Config Change

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

### Loki Promtail Config Change

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
