---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/install/install-agent-kubernetes/
- /docs/grafana-cloud/send-data/agent/static/set-up/install/install-agent-kubernetes/
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-kubernetes/
description: Learn how to deploy Grafana Agent in static mode on Kubernetes
menuTitle: Kubernetes
title: Deploy Grafana Agent in static mode on Kubernetes
weight: 300
---

# Deploy Grafana Agent in static mode on Kubernetes

You can use the Helm chart for Grafana Agent to deploy Grafana Agent in static mode on Kubernetes.

## Before you begin

* Install [Helm][] on your computer.
* Configure a Kubernetes cluster that you can use for Grafana Agent.
* Configure your local Kubernetes context to point to the cluster.

[Helm]: https://helm.sh

## Deploy

{{< admonition type="note" >}}
These instructions show you how to install the generic [Helm chart](https://github.com/grafana/agent/tree/main/operations/helm/charts/grafana-agent) for Grafana Agent.
You can deploy Grafana Agent in static mode or flow mode. The Helm chart deploys flow mode by default.
{{< /admonition >}}

To deploy Grafana Agent in static mode on Kubernetes using Helm, run the following commands in a terminal window:

1. Add the Grafana Helm chart repository:

   ```shell
   helm repo add grafana https://grafana.github.io/helm-charts
   ```

1. Update the Grafana Helm chart repository:

   ```shell
   helm repo update
   ```

1. Install Grafana Agent in static mode:

   ```shell
   helm install <RELEASE_NAME> grafana/grafana-agent --set agent.mode=static
   ```

   Replace the following:

   -  _`<RELEASE_NAME>`_: The name to use for your Grafana Agent installation, such as `grafana-agent`.

   {{< admonition type="warning" >}}
   Always pass `--set agent.mode=static` in `helm install` or `helm upgrade` commands to ensure Grafana Agent gets installed in static mode.
   Alternatively, set `agent.mode` to `static` in your values.yaml file.
   {{< /admonition >}}

For more information on the Grafana Agent Helm chart, refer to the Helm chart documentation on [Artifact Hub][].

[Artifact Hub]: https://artifacthub.io/packages/helm/grafana/grafana-agent

