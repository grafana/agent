---
aliases:
- ../../install/kubernetes/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/kubernetes/
description: Learn how to deploy Grafana Agent in flow mode on Kubernetes
menuTitle: Kubernetes
title: Deploy Grafana Agent in flow mode on Kubernetes
weight: 200
---

# Deploy Grafana Agent in flow mode on Kubernetes

Grafana Agent can be deployed on Kubernetes by using the Helm chart for Grafana Agent.

## Before you begin

* Install [Helm][] on your computer.
* Configure a Kubernetes cluster that you can use for Grafana Agent.
* Configure your local Kubernetes context to point to the cluster.

[Helm]: https://helm.sh

## Deploy

{{% admonition type="note" %}}
These instructions show you how to install the generic [Helm chart](https://github.com/grafana/agent/tree/main/operations/helm/charts/grafana-agent) for Grafana
Agent. You can deploy Grafana Agent either in static mode or flow mode. The Helm chart deploys Grafana Agent in flow mode by default.
{{% /admonition %}}

To deploy Grafana Agent on Kubernetes using Helm, run the following commands in a terminal window:

1. Add the Grafana Helm chart repository:

   ```shell
   helm repo add grafana https://grafana.github.io/helm-charts
   ```

1. Update the Grafana Helm chart repository:

   ```shell
   helm repo update
   ```

1. Install Grafana Agent:

   ```shell
   helm install RELEASE_NAME grafana/grafana-agent
   ```

   Replace `RELEASE_NAME` with a name to use for your Grafana Agent
   installation, such as `grafana-agent-flow`.

For more information on the Grafana Agent Helm chart, refer to the Helm chart documentation on [Artifact Hub][].

[Artifact Hub]: https://artifacthub.io/packages/helm/grafana/grafana-agent

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent#linux" >}})
- [Configure Grafana Agent]({{< relref "../configure/configure-linux" >}})
