# Getting Started

## Docker-Compose Example

The quickest way to try out the Agent with a full Cortex, Grafana, and Agent
stack is to check out this repository's [Docker-Compose Example](../example/docker-compose/README.md):

```
# Clone the Grafana Cloud Agent repository
git clone https://github.com/grafana/agent.git
cd agent/example/docker-compose
docker-compose up -d

# Navigate to localhost:3000 in your browser
```

## k3d Example

A more advanced [Kubernetes example](../example/k3d/README.md) using a local
cluster and Tanka is provided to deploy the Agent "normally" alongside a
[Scraping Service](./scraping-service.md) deployment:

```
# Clone the Grafana Cloud Agent repository
git clone https://github.com/grafana/agent.git
cd agent/example/k3d
./scripts/create.bash

# Wait a little bit, 5-10 seconds
./scripts/merge_k3d.bash
tk apply ./environment

# Navigate to localhost:30080 in your browser
```

## Installing

Currently, there are five ways to install the agent:

1. Use our Docker container
2. Use the Kubernetes install script (_recommended basic_)
3. Use the Kubernetes manifest
4. Installing the static binaries locally
5. Using Grafana Labs' official Tanka configs (_recommended advanced_)

### Docker Container

```
docker pull grafana/agent:v0.4.0
```

### Kubernetes Install Script

Running this script will automatically download and apply our recommended
Grafana Cloud Agent Kubernetes deployment manifest (requires `envsubst` (GNU gettext)):

> **Warning**: Always verify scripts from the internet before running them.

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/v0.4.0/production/kubernetes/install.sh)" | kubectl apply -f -
```

### Kubernetes Manifest

If you wish to manually modify the Kubernetes manifest before deploying it
yourself, you can do so by downloading the [`agent.yaml` file](/production/kubernetes/agent.yaml).

### Installing Locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.

### Tanka

We provide [Tanka](https://tanka.dev) configurations in our [`production/`](/production/tanka/grafana-agent) directory.

## Creating a Config File

The Grafana Cloud Agent can be configured with **integrations** and a
**Prometheus-like** config. Both may coexist in the same configuration file for
the Agent.

Integrations are recommended for first-time users of monitoring or Prometheus;
users that have more experience with Prometheus or already have an existing
Prometheus config file should use the Prometheus-like config.

### Integrations

**Integrations** are subsystems that collect metrics for you. For example, the
`agent` integration collects metrics from that running instance of the Grafana
Cloud Agent. The `node_exporter` integration will collect metrics from the Linux
machine that the Grafana Cloud Agent is running on.

```yaml
prometheus:
  wal_directory: /tmp/wal

integrations:
  agent:
    enabled: true
  prometheus_remote_write:
    - url: http://localhost:9009/api/prom/push
```

In this example, we first must configure the `wal_directory` which is used to
store metrics in a Write-Ahead Log. This is required, but ensures that samples
will be resent in case of failure (e.g., network issues, machine reboot).

Then, the `integrations` are configured. In this example, just the `agent`
integration is enabled. Finally, `prometheus_remote_write` is configured with a
location to send metrics. You will have to replace this URL with the appropriate
URL for your `remote_write` system (such as a Grafana Cloud Hosted Prometheus
instance).

When the Agent is run with this file, it will collect metrics from itself and
send those metrics to the `remote_write` endpoint. All metrics will have (by
default) an `agent_hostname` label equal to the hostname of the machine the
Agent is running on. This label helps to uniquly identify the source of metrics
if you run multiple Agent processes across multiple machines.

Full configuration options can be found in the
[configuration reference](./configuration-reference.md).

## Prometheus-like Config/Migrating from Prometheus

The Prometheus-like Config is useful for those migrating from Prometheus and
those who want to scrape metrics from something that currently does not have an
associated integration.

To migrate from an existing Prometheus config, use this Agent config as a
template and copy and paste subsections from your existing Prometheus config
into it:

```yaml
prometheus:
  global:
  # PASTE PROMETHEUS global SECTION HERE
  configs:
    - name: agent
      # Leave this as false for a Prometheus-like agent process
      host_filter: false
      scrape_configs:
        # PASTE scrape_configs SECTION HERE
      remote_write:
        # PASTE remote_write SECTION HERE
```

For example, this configuration file configures the Grafana Cloud Agent to
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
      host_filter: false
      scrape_configs:
        - job_name: agent
          static_configs:
            - targets: ['127.0.0.1:12345']
      remote_write:
        - url: http://localhost:9009/api/prom/push
```

Like with integrations, full configuration options can be found in the
[configuration reference](./configuration-reference.md).

## Running

If you've installed the agent with Kubernetes, it's already running! The
following sections below describe running the agent in environments that need
extra steps.

### Docker Container

Copy the following block below, replacing `/tmp/agent` with the host directory
where you want to store the agent WAL and `/path/to/config.yaml` with the full
path of your Agent's YAML configuration file.

```
docker run \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:v0.4.0
```

### Locally

This section is only relavant if you installed the static binary of the
Agent. We do not yet provide system packages or configurations to run the Agent
as a daemon process.

To override the default flags passed to the container, add the following flags
to the end of the `docker run` command:

- `--config.file=path/to/agent.yaml`, replacing the argument with the full path
  to your Agent's YAML configuration file.

- `--prometheus.wal-directory=/tmp/agent/data`, replacing `/tmp/agent/data` with
  the directory you wish to use for storing data. Note that `/tmp` may get
  deleted by most operating systems after a reboot.

Note that using paths on your host machine must be exposed to the Docker
container through a bind mount for the flags to work properly.
