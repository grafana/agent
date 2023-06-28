---
description: Learn how to install Grafana Agent in flow mode on Kubernetes
title: Install Grafana Agent in flow mode on Kubernetes
menuTitle: Kubernetes
weight: 200
aliases:
 - ../../install/kubernetes/
---

# Install Grafana Agent in flow mode on Kubernetes

Grafana Agent can be installed on Kubernetes by using the Helm chart for Grafana Agent.

## Before you begin

* Ensure that you have [Helm][] installed on your system.
* Ensure that you have a Kubernetes cluster to deploy Grafana Agent to.
* Ensure that your local Kubernetes context is configured to point to the
  correct cluster.

[Helm]: https://helm.sh

## Install

{{% admonition type="note" %}}
These instructions install the generic [Helm chart](https://github.com/grafana/agent/tree/main/operations/helm/charts/grafana-agent) for Grafana
Agent, which can deploy Grafana Agent either in static mode or flow mode.
The Helm chart deploys Grafana Agent in flow mode by default.
{{% /admonition %}}

To install Grafana Agent on Kubernetes using Helm, perform the following
steps in a terminal:

1. Add the Grafana Helm chart repository:

   ```shell
   helm repo add grafana https://grafana.github.io/helm-charts
   ```

1. Ensure that the Grafana Helm chart repository is up to date:

   ```shell
   helm repo update
   ```

1. Install the Grafana Agent Helm chart:

   ```shell
   helm install RELEASE_NAME grafana/grafana-agent
   ```

   Replace `RELEASE_NAME` with a name to use for your Grafana Agent
   installation, such as `grafana-agent-flow`.

For more information on the Grafana Agent Helm chart, refer to the Helm chart
on [Artifact Hub][].

[Artifact Hub]: https://artifacthub.io/packages/helm/grafana/grafana-agent
