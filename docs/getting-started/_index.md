+++
title = Getting started with Grafana Agent
weight = 100
+++

# Getting started with Grafana Agent

This guide helps users get started with the Grafana Agent. For getting started
with the Grafana Agent Operator, please refer to the Operator-specific
[documentation](../operator/_index.md).

Currently, there are six ways to install the agent:

- Use our Docker container
- Use the Kubernetes install script (_recommended basic_)
- Use the Kubernetes manifest
- Installing the static binaries locally
- Using Grafana Labs' official Tanka configs (_recommended advanced_)
- Using the [Windows Installer]({{< relref "./install-agent-on-windows.md" >}})

## Docker container

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

## Kubernetes install script

Running this script automatically downloads and applies our recommended
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

## Kubernetes manifest

If you wish to manually modify the Kubernetes manifest before deploying it
yourself, you can do so by downloading the [`agent.yaml` file](../../production/kubernetes/agent.yaml).

## Install locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.
These releases contain the plain binary alongside system packages for Windows,
Red Hat, and Debian.

## Tanka

We provide [Tanka](https://tanka.dev) configurations in our [`production/`](https://github.com/grafana/agent/tree/main/production/tanka/grafana-agent) directory.

