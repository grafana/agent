+++
title = "Upgrade guide"
weight = 200
+++

# Upgrade guide

This guide describes all breaking changes that have happened in prior
releases and how to migrate to newer versions.

## Unreleased Changes

These changes will come in a future version.

## v0.19.0

### Traces: Deprecation of "tempo" in config and metrics. (Deprecation)

The term `tempo` in the config has been deprecated of favor of `traces`. This
change is to make intent clearer.

Example old config:

```yaml
tempo:
  configs:
    - name: default
      receivers:
        jaeger:
          protocols:
            thrift_http:
```

Example of new config:
```yaml
traces:
  configs:
    - name: default
      receivers:
        jaeger:
          protocols:
            thrift_http:
```

Any tempo metrics have been renamed from `tempo_*` to `traces_*`.


### Tempo: split grouping by trace from tail sampling config (Breaking change)

Load balancing traces between agent instances has been moved from an embedded
functionality in tail sampling to its own configuration block.
This is done due to more processor benefiting from receiving consistently
receiving all spans for a trace in the same agent to be processed, such as
service graphs.

As a consequence, `tail_sampling.load_balancing` has been deprecated in favor of
a `load_balancing` block. Also, `port` has been renamed to `receiver_port` and
moved to the new `load_balancing` block.

Example old config:

```yaml
tail_sampling:
  policies:
    - always_sample:
  port: 4318
  load_balancing:
    exporter:
      insecure: true
    resolver:
      dns:
        hostname: agent
        port: 4318
```

Example new config:

```yaml
tail_sampling:
  policies:
    - always_sample:
load_balancing:
  exporter:
    insecure: true
  resolver:
    dns:
      hostname: agent
      port: 4318
  receiver_port: 4318
```

### Operator: Rename of Prometheus to Metrics (Breaking change)

As a part of the deprecation of "Prometheus," all Operator CRDs and fields with
"Prometheus" in the name have changed to "Metrics."

This includes:

- The `PrometheusInstance` CRD is now `MetricsInstance` (referenced by
  `metricsinstances` and not `metrics-instances` within ClusterRoles).
- The `Prometheus` field of the `GrafanaAgent` resource is now `Metrics`
- `PrometheusExternalLabelName` is now `MetricsExternalLabelName`

This is a hard breaking change, and all fields must change accordingly for the
operator to continue working.

To do a zero-downtime upgrade of the Operator when there is a breaking change,
refer to the new `agentctl operator-detatch` command: this will iterate through
all of your objects and remove any OwnerReferences to a CRD, allowing you to
delete your Operator CRDs or CRs.

### Operator: Rename of CRD paths (Breaking change)

`prometheus-instances` and `grafana-agents` have been renamed to
`metricsinstances` and `grafanaagents` respectively. This is to remain
consistent with how Kubernetes names multi-word objects.

As a result, you will need to update your ClusterRoles to change the path of
resources.

To do a zero-downtime upgrade of the Operator when there is a breaking change,
refer to the new `agentctl operator-detatch` command: this will iterate through
all of your objects and remove any OwnerReferences to a CRD, allowing you to
delete your Operator CRDs or CRs.


Example old ClusterRole:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-operator
rules:
- apiGroups: [monitoring.grafana.com]
  resources:
  - grafana-agents
  - prometheus-instances
  verbs: [get, list, watch]
```

Example new ClusterRole:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-operator
rules:
- apiGroups: [monitoring.grafana.com]
  resources:
  - grafanaagents
  - metricsinstances
  verbs: [get, list, watch]
```

### Metrics: Deprecation of "prometheus" in config. (Deprecation)

The term `prometheus` in the config has been deprecated of favor of `metrics`. This
change is to make it clearer when referring to Prometheus or another
Prometheus-like database, and configuration of Grafana Agent to send metrics to
one of those systems.

Old configs will continue to work until it is fully deprecated. To migrate your
config, change the `prometheus` key to `metrics`.

Example old config:

```yaml
prometheus:
  configs:
    - name: default
      host_filter: false
      scrape_configs:
        - job_name: local_scrape
          static_configs:
            - targets: ['127.0.0.1:12345']
              labels:
                cluster: 'localhost'
      remote_write:
        - url: http://localhost:9009/api/prom/push
```

Example new config:

```yaml
metrics:
  configs:
    - name: default
      host_filter: false
      scrape_configs:
        - job_name: local_scrape
          static_configs:
            - targets: ['127.0.0.1:12345']
              labels:
                cluster: 'localhost'
      remote_write:
        - url: http://localhost:9009/api/prom/push
```

### Tempo: prom_instance rename (Breaking change)

As part of `prometheus` being renamed to `metrics`, the spanmetrics
`prom_instance` field has been renamed to `metrics_instance`. This is a breaking
change, and the old name will no longer work.

Example old config:

```yaml
tempo:
  configs:
  - name: default
    spanmetrics:
      prom_instance: default
```

Example new config:

```yaml
tempo:
  configs:
  - name: default
    spanmetrics:
      metrics_instance: default
```

### Logs: Deprecation of "loki" in config. (Deprecation)

The term `loki` in the config has been deprecated of favor of `logs`. This
change is to make it clearer when referring to Grafana Loki, and
configuration of Grafana Agent to send logs to Grafana Loki.

Old configs will continue to work until it is fully deprecated. To migrate your
config, change the `loki` key to `logs`.

Example old config:

```yaml
loki:
  positions_directory: /tmp/loki-positions
  configs:
  - name: default
    clients:
      - url: http://localhost:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
      - targets: ['localhost']
        labels:
          job: varlogs
          __path__: /var/log/*log
```

Example new config:

```yaml
logs:
  positions_directory: /tmp/loki-positions
  configs:
  - name: default
    clients:
      - url: http://localhost:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
      - targets: ['localhost']
        labels:
          job: varlogs
          __path__: /var/log/*log
```

#### Tempo: Deprecation of "loki" in config. (Deprecation)

As part of the `loki` to `logs` rename, parts of the automatic_logging component
in Tempo have been updated to refer to `logs_instance` instead.

Old configurations using `loki_name`, `loki_tag`, or `backend: loki` will
continue to work until the `loki` terminology is fully deprecated.

Example old config:

```yaml
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: loki
      loki_name: default
      spans: true
      processes: true
      roots: true
    overrides:
      loki_tag: tempo
```

Example new config:

```yaml
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: logs_instance
      logs_instance_name: default
      spans: true
      processes: true
      roots: true
    overrides:
      logs_instance_tag: tempo
```

## v0.18.0

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
