# Getting Started

## Docker-Compose Example

The quickest way to try out the Agent with a full Cortex, Grafana, and Agent
stack is to check out this repository's [Docker-Compose Example](../example/docker-compose/README.md):

```
# Clone the Grafana Agent repository
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
# Clone the Grafana Agent repository
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
docker pull grafana/agent:v0.13.0
```

### Kubernetes Install Script

Running this script will automatically download and apply our recommended
Grafana Agent Kubernetes deployment manifests (requires `envsubst` (GNU gettext)).
Two manifests will be installed: one for collecting metrics, and the other for
collecting logs. You will be prompted for input for each manifest that is
applied.

> **Warning**: Always verify scripts from the internet before running them.

```
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install.sh)" | kubectl -ndefault apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-loki.sh)" | kubectl apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-tempo.sh)" | kubectl apply -f -
```
**Note:** For the above script to scrape your pods, they must conform to the following rules:

1. The pod must _not_ have an annotation matching `prometheus.io/scrape: "false"` (this wouldn't be there unless you explicitly add it or if you deploy a Helm chart that has it).
2. The pod _must_ have a port with a name ending in `-metrics`. This is the port that will be scraped by the Agent. A lot of people using Helm struggle with this, since Helm charts don't usually follow this. You would need to add a new scrape config to scrape helm charts or find a way to tweak the Helm chart to follow this rules.
3. The pod _must_ have a label named name with any non-empty value. Helm usually lets you add extra labels, so this is less of a problem for Helm users.
4. The pod must currently be running. (i.e., Kubernetes must not report it having a phase of Succeeded or Failed).

### Kubernetes Manifest

If you wish to manually modify the Kubernetes manifest before deploying it
yourself, you can do so by downloading the [`agent.yaml` file](/production/kubernetes/agent.yaml).

### Installing Locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.

### Tanka

We provide [Tanka](https://tanka.dev) configurations in our [`production/`](/production/tanka/grafana-agent) directory.

## Creating a Config File

The Grafana Agent supports configuring **integrations**, a
**Prometheus-like** config, and a **Loki** config. All may coexist together
within the same configuration file for the Agent.

Integrations are recommended for first-time users of observability platforms,
especially newcomers to Prometheus. Users with more experience with Prometheus
or users that already have an existing Prometheus config file should use the
Prometheus-like config.

### Integrations

**Integrations** are subsystems that collect metrics for you. For example, the
`agent` integration collects metrics from that running instance of the Grafana
Agent. The `node_exporter` integration will collect metrics from the Linux
machine that the Grafana Agent is running on.

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
Agent is running on. This label helps to uniquely identify the source of metrics
if you run multiple Agent processes across multiple machines.

Full configuration options can be found in the
[configuration reference](./configuration-reference.md).

### Prometheus-like Config/Migrating from Prometheus

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
server:
  # AGENT SERVER SETTINGS

prometheus:
  # AGENT PROMETHEUS SETTINGS

loki:
  configs:
  - name: default
    # PASTE YOUR PROMTAIL CONFIG INSIDE OF HERE

tempo:
  # AGENT TEMPO SETTINGS

integrations:
  # AGENT INTEGRATIONS SETTINGS
```

Here is an example full config file, using integrations,
Prometheus, Loki, and Tempo:

```yaml
server:
  log_level: info
  http_listen_port: 12345

prometheus:
  global:
    scrape_interval: 5s
  configs:
    - name: default
      scrape_configs:
        - job_name: agent
          static_configs:
            - targets: ['127.0.0.1:12345']
      remote_write:
        - url: http://localhost:9009/api/prom/push

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
```

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
  -v /path/to/config.yaml:/etc/agent/agent.yaml \
  grafana/agent:v0.13.0
```

### Locally

This section is only relevant if you installed the static binary of the
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
