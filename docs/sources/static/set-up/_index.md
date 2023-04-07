---
title: Set up Grafana Agent
weight: 100
aliases:
- ../set-up/
---

# Set up Grafana Agent

## Overview

If this is your first time using Grafana Agent, use one of the installation options to install Grafana Agent based on the platform you are using. Alternatively, use the quick start guides to help you with the specifics of sending metrics, logs, and traces to the Grafana Stack or Grafana Cloud.

If you have already installed Grafana Agent on your machine, you can jump to the [Configure Grafana Agent]({{< relref "../configuration/_index.md" >}}) section.

To get started with Grafana Agent Operator, refer to the Operator-specific
[documentation](../operator/).

## Installation options

Grafana Agent is currently distributed in plain binary form, Docker container images, a Windows installer, a Homebrew package, and a Kubernetes install script.

The following architectures receive active support.

 - macOS: Intel Mac or Apple Silicon
 - Windows: A x64 machine
 - Linux: AMD64 or ARM64 machines
 - FreeBSD: A AMD64 machine

In addition, best-effort support is provided for Linux: ppc64le.

Choose from the following platforms and installation options according to which suits your use case best.

### Kubernetes

Deploy Kubernetes manifests from the [`kubernetes` directory](https://github.com/grafana/agent/tree/main/production/kubernetes).
You can manually modify the Kubernetes manifests by downloading them. These manifests do not include Grafana Agent configuration files.

For sample configuration files, refer to the Grafana Cloud Kubernetes quick start guide: https://grafana.com/docs/grafana-cloud/kubernetes/agent-k8s/.

Advanced users can use the Grafana Agent Operator to deploy the Grafana Agent on Kubernetes.

### Docker

Refer to [Install Grafana Agent on Docker]({{< relref "./install-agent-docker.md" >}}).

### Linux

Refer to [Install Grafana Agent on Linux]({{< relref "./install-agent-linux.md" >}}).

### Windows

Refer to [Install Grafana Agent on Windows]({{< relref "./install-agent-on-windows.md" >}}).

### Binary

Refer to [Install the Grafana Agent binary]({{< relref "./install-agent-binary.md" >}}).

### macOS

Refer to [Install Grafana Agent on macOS]({{< relref "./install-agent-macos.md" >}}).

### Grafana Cloud

Use the Grafana Agent [Kubernetes quickstarts](https://grafana.com/docs/grafana-cloud/kubernetes/agent-k8s/) or follow instructions for installing the Grafana Agent in the [Walkthrough](https://grafana.com/docs/grafana-cloud/quickstart/agent_linuxnode/).

### Tanka

For more information, refer to the [Tanka](https://tanka.dev) configurations in our [`production/`](https://github.com/grafana/agent/tree/main/production/tanka/grafana-agent) directory.
