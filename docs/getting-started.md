# Getting Started

## Docker-Compose Example

The quickest way to try out the Agent with a full Cortex, Grafana, and Agent
stack is to check out this repository's [Docker-Compose Example](/example):

```
# Clone the Grafana Cloud Agent repository
git clone https://github.com/grafana/agent.git
cd agent/example
docker-compose up -d

# Navigate to localhost:3000 in your browser
```

## Installing

Currently, there are five ways to install the agent:

1. Use our Docker container
2. Use the Kubernetes install script
3. Use the Kubernetes manifest
4. Installing the static binaries locally
5. Using Grafana Labs' official Tanka configs

### Docker Container

```
docker pull grafana/agent:v0.1.0
```

### Kubernetes Install Script

Running this script will automatically download and apply our recommended
Grafana Cloud Agent Kubernetes deployment manifest:

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/v0.1.0/production/kubernetes/install.sh)" | kubectl apply -f -
```

### Kubernetes Manifest

If you wish to manually modify the Kubernetes manifest before deploying it
yourself, you can do so by downloading [this](/production/kubernetes/agent.yaml) file.

### Installing Locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.

### Tanka

We provide Tanka configurations in our [production](/production/tanka/grafana-agent) directory.

## Migrating from Prometheus

Migrating from Prometheus is relatively painless, requiring just copying and
pasting sections from the Prometheus configuration file into the Agent
configuration file. Use this Agent config as a template:

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
  --entrypoint "/bin/agent -config.file=/etc/agent-config/agent.yaml -prometheus.wal-directory=/etc/agent/data"
  grafana/agent:v0.1.0
```

### Locally

This section is only relavant if you have installed the static binary of the
Agent. We do not yet provide system packages or configurations to run the Agent
as a daemon process.

To run the agent, you need to pass the following flags to it:

- `-config.file=path/to/agent.yaml`, replacing the argument with the full path
   to your Agent's YAML configuration file.

- `-prometheus.wal-directory=/tmp/agent/data`, replacing `/tmp/agent/data` with
  the directory you wish to use for storing data. Note that `/tmp` may get
  deleted by most operating systems after a reboot.

