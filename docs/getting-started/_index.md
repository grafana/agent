+++
title = Getting started with Grafana Agent
weight = 100
+++

# Getting started with Grafana Agent

This guide helps users get started with the Grafana Agent. For getting started
with the Grafana Agent Operator, please refer to the Operator-specific
[documentation](../operator/_index.md).

## Installation methods

Currently, there are six ways to install the agent:

- Use our Docker container
- Use the Kubernetes install script (_recommended basic_)
- Use the Kubernetes manifest
- Installing the static binaries locally
- Using Grafana Labs' official Tanka configs (_recommended advanced_)
- Using the [Windows Installer](./install-agent-on-windows.md)

### Docker container

```
docker run \
  -v /tmp/agent:/etc/agent/data \
  -v /path/to/config.yaml:/etc/agent/agent.yaml \
  grafana/agent:v0.16.1
```

Replace `/tmp/agent` with the folder you wish to store WAL data in. WAL data is
where metrics are stored before they are sent to Prometheus. Old WAL data is
cleaned up every hour, and will be used for recovering if the process happens to
crash.

To override the default flags passed to the container, add the following flags
to the end of the `docker run` command:

- `--config.file=path/to/agent.yaml`, replacing the argument with the full path
  to your Agent's YAML configuration file.

- `--prometheus.wal-directory=/tmp/agent/data`, replacing `/tmp/agent/data` with
  the directory you wish to use for storing data. Note that `/tmp` may get
  deleted by most operating systems after a reboot.

Note that using paths on your host machine must be exposed to the Docker
container through a bind mount for the flags to work properly.

### Kubernetes Install Script

Running this script will automatically download and apply our recommended
Grafana Agent Kubernetes deployment manifests (requires `envsubst` (GNU gettext)).
Two manifests will be installed: one for collecting metrics, and the other for
collecting logs. You will be prompted for input for each manifest that is
applied.

> **Warning:** Always verify scripts from the internet before running them.

```
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install.sh)" | kubectl -ndefault apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-loki.sh)" | kubectl apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-tempo.sh)" | kubectl apply -f -
```

**Note:** For the above script to scrape your pods, they must conform to the following rules:

- The pod must _not_ have an annotation matching `prometheus.io/scrape: "false"` (this wouldn't be there unless you explicitly add it or if you deploy a Helm chart that has it).
- The pod _must_ have a port with a name ending in `-metrics`. This is the port that will be scraped by the Agent. A lot of people using Helm struggle with this, since Helm charts don't usually follow this. You would need to add a new scrape config to scrape helm charts or find a way to tweak the Helm chart to follow this rules.
- The pod _must_ have a label named name with any non-empty value. Helm usually lets you add extra labels, so this is less of a problem for Helm users.
- The pod must currently be running. (i.e., Kubernetes must not report it having a phase of Succeeded or Failed).

### Kubernetes Manifest

If you wish to manually modify the Kubernetes manifest before deploying it
yourself, you can do so by downloading the [`agent.yaml` file](../../production/kubernetes/agent.yaml).

### Install locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.
These releases contain the plain binary alongside system packages for Windows,
Red Hat, and Debian.

### Tanka

We provide [Tanka](https://tanka.dev) configurations in our [`production/`](../../production/tanka/grafana-agent) directory.

## Create a Config File

The Grafana Agent supports configuring multiple independent "subsystems." Each
subsystem helps you collect data for a specific type of telemetry.

* The **Prometheus** subsystem allows you collect metrics to send to Prometheus.
* The **Loki** subsystem allows you to collect logs to send to Grafana Loki.
* The **Tempo** subsystem allows you to collect spans to send to Grafana Tempo.
* The **Integrations** subsystem allows you to collect metrics for common
  applications, such as MySQL.

Integrations are recommended for first-time users of observability platforms,
especially newcomers to Prometheus. Users with more experience with Prometheus
or users that already have an existing Prometheus config file can configure
the Prometheus subsystem manually.

### Integrations

**Integrations** are individual features that collect metrics for you. For
example, the `agent` integration collects metrics from that running instance of
the Grafana Agent. The `node_exporter` integration will collect metrics from the
Linux machine that the Grafana Agent is running on.

```yaml
prometheus:
  wal_directory: /tmp/wal
  global:
    remote_write:
      - url: http://localhost:9009/api/prom/push

integrations:
  agent:
    enabled: true
```

In this example, we first must configure the `wal_directory` which is used to
store metrics in a Write-Ahead Log. This is required, but ensures that samples
will be resent in case of failure (e.g., network issues, machine reboot). We
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

Full configuration options can be found in the
[configuration reference](../configuration/_index.md).

### Prometheus Config/Migrating from Prometheus

The Prometheus subsystem config is useful for those migrating from Prometheus
and those who want to scrape metrics from something that currently does not have
an associated integration.

To migrate from an existing Prometheus config, use this Agent config as a
template and copy and paste subsections from your existing Prometheus config
into it:

```yaml
prometheus:
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
  http_listen_port: 12345

prometheus:
  global:
    scrape_interval: 5s
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
[configuration](../configuration/_index.md).

### Loki Config/Migrating from Promtail

The Loki Config allows for collecting logs to send to a Loki API. Users that are
familiar with Promtail will notice that the Loki config for the Agent matches
their existing Promtail config with the following exceptions:

1. The deprecated field `client` is not present
2. The `server` field is not present

To migrate from an existing Promtail config, make sure you are using `clients`
instead of `client` and remove the `server` block if present. Then paste your
Promtail config into the Agent config file inside of a `loki` section:

```yaml
loki:
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
  http_listen_port: 12345

prometheus:
  global:
    scrape_interval: 5s
    remote_write:
      - url: http://localhost:9009/api/prom/push
  configs:
    - name: default
      scrape_configs:
        - job_name: agent
          static_configs:
            - targets: ['127.0.0.1:12345']

loki:
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

tempo:
  configs:
  - name: default
    receivers:
      jaeger:
        protocols:
          grpc: # listens on the default jaeger grpc port: 14250
    remote_write:
      - endpoint: localhost:55680
        insecure: true  # only add this if TLS is not required
        queue:
          retry_on_failure: true
    batch:
      timeout: 5s
      send_batch_size: 100

integrations:
  node_exporter:
    enabled: true
```
