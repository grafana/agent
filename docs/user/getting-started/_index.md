+++
title = "Getting started with Grafana Agent"
weight = 100
+++

# Getting started with Grafana Agent

This guide helps users get started with the Grafana Agent. For getting started
with the Grafana Agent Operator, please refer to the Operator-specific
[documentation](../operator/).

Currently, there are six ways to install the agent:

- Use our Docker container
- Use the Kubernetes manifests directly
- Use the Kubernetes manifests along with the [Grafana Cloud Kubernetes Quickstart Guides](#grafana-cloud-kubernetes-quickstart-guides)
- Installing the static binaries locally
- Using Grafana Labs' official Tanka configs (_recommended advanced_)
- Using the [Windows Installer]({{< relref "./install-agent-on-windows.md" >}})

See the list of [Community Projects](#community-projects) for the community-driven ecosystem around the Grafana Agent.

## Docker container

```
docker run \
  -v /tmp/agent:/etc/agent/data \
  -v /path/to/config.yaml:/etc/agent/agent.yaml \
  grafana/agent:v0.24.1
```

Replace `/tmp/agent` with the folder you wish to store WAL data in. WAL data is
where metrics are stored before they are sent to Prometheus. Old WAL data is
cleaned up every hour, and will be used for recovering if the process happens to
crash.

To override the default flags passed to the container, add the following flags
to the end of the `docker run` command:

- `--config.file=path/to/agent.yaml`, replacing the argument with the full path
  to your Agent's YAML configuration file.

- `--metrics.wal-directory=/tmp/agent/data`, replacing `/tmp/agent/data` with
  the directory you wish to use for storing data. Note that `/tmp` may get
  deleted by most operating systems after a reboot.

Note that using paths on your host machine must be exposed to the Docker
container through a bind mount for the flags to work properly.

## Kubernetes manifests

If you wish to manually modify the Kubernetes manifests before deploying them, you can do so by downloading them from the [`kubernetes` directory](https://github.com/grafana/agent/tree/main/production/kubernetes). Note that these manifests do not include Agent configuration files. For sample configuration, please see the Grafana Cloud Kubernetes quickstarts.

## Grafana Cloud kubernetes quickstart guides

These guides help you get up and running with the Agent and Grafana Cloud, and include sample ConfigMaps.

You can find them in the [Grafana Cloud documentation](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/)

## Install locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.
These releases contain the plain binary alongside system packages for Windows,
Red Hat, and Debian.

## Tanka

We provide [Tanka](https://tanka.dev) configurations in our [`production/`](https://github.com/grafana/agent/tree/main/production/tanka/grafana-agent) directory.

## Community Projects

Below is a list of community lead projects for working with Grafana Agent. These projects are not maintained or supported by Grafana Labs.

### Helm (Kubernetes Deployment)

A publically available release of a Grafana Agent Helm chart is maintained [here](https://github.com/DandyDeveloper/charts/tree/master/charts/grafana-agent). Contributions and improvements are welcomed. Full details on rolling out and supported options can be found in the [readme](https://github.com/DandyDeveloper/charts/blob/master/charts/grafana-agent/README.md).

This *does not* require the Grafana Agent Operator to rollout / deploy.

### Juju (Charmed Operator)

The [grafana-agent-k8s](https://github.com/canonical/grafana-agent-operator) charmed operator runs with [Juju](https://juju.is) the Grafana Agent on Kubernetes.
The Grafana Agent charmed operator is designed to work with the [Logs, Metrics and Alerts](https://juju.is/docs/lma2) observability stack.
