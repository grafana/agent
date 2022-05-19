---
aliases:
- /docs/agent/latest/set-up/
title: Set up Grafana Agent
weight: 100
---

# Set up Grafana Agent

## Overview
If this is your first time using Grafana Agent, use one of the installation options to install Grafana Agent based on the platform you are using. Alternatively, use the quick start guides to help you with the specifics of sending metrics, logs, and traces to the LGTM (Loki, Grafana, Tempo, Mimir) Stack or Grafana Cloud.

If you have already installed Grafana Agent on your machine, you can jump to the Configure Grafana Agent section.

To get started with Grafana Agent Operator, refer to the Operator-specific
[documentation](../operator/).

## Installation options

Grafana Agent is currently distributed in plain binary form, Docker container images, a Windows installer, a Homebrew package, and a Kubernetes install script. 

The following architectures receive active support.

 - macOS: Intel Mac or Apple Silicon 
 - Windows: A x64 machine 
 - Linux: AMD64, ARM64, ARMv6, or ARMv7 machines
 - FreeBSD: A AMD64 machine 

In addition, best-effort support is provided for Linux: ppc64le.

Choose from the following platforms and installation options according to which suits your use case best.

### Kubernetes

Deploy Kubernetes manifests from the [`kubernetes` directory](https://github.com/grafana/agent/tree/main/production/kubernetes).
You can manually modify the Kubernetes manifests by downloading them. These manifests do not include Grafana Agent configuration files. 

For sample configuration files, refer to the Grafana Cloud Kubernetes quick start guide: https://grafana.com/docs/grafana-cloud/kubernetes/agent-k8s/.

Advanced users can use Grafana Agent Operator to deploy the Grafana Agent on Kubernetes.

### Docker

Refer to [Install Grafana Agent on Docker]({{< relref "./install-agent-docker.md" >}})

### Windows

Refer to [Install Grafana Agent on Windows]({{< relref "./install-agent-on-windows.md" >}})

### Install locally

Our [Releases](https://github.com/grafana/agent/releases) page contains
instructions for downloading static binaries that are published with every release.
These releases contain the plain binary alongside system packages for Windows,
Red Hat, and Debian.

### Tanka

We provide [Tanka](https://tanka.dev) configurations in our [`production/`](https://github.com/grafana/agent/tree/main/production/tanka/grafana-agent) directory.

### Community Projects

Below is a list of community lead projects for working with Grafana Agent. These projects are not maintained or supported by Grafana Labs.

#### Helm (Kubernetes Deployment)

A publically available release of a Grafana Agent Helm chart is maintained [here](https://github.com/DandyDeveloper/charts/tree/master/charts/grafana-agent). Contributions and improvements are welcomed. Full details on rolling out and supported options can be found in the [readme](https://github.com/DandyDeveloper/charts/blob/master/charts/grafana-agent/README.md).

This *does not* require the Grafana Agent Operator to rollout / deploy.

#### Juju (Charmed Operator)

The [grafana-agent-k8s](https://github.com/canonical/grafana-agent-operator) charmed operator runs with [Juju](https://juju.is) the Grafana Agent on Kubernetes.
The Grafana Agent charmed operator is designed to work with the [Logs, Metrics and Alerts](https://juju.is/docs/lma2) observability stack.
