---
aliases:
- /docs/grafana-cloud/agent/operator/helm-getting-started/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/helm-getting-started/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/helm-getting-started/
- /docs/grafana-cloud/send-data/agent/operator/helm-getting-started/
canonical: https://grafana.com/docs/agent/latest/operator/helm-getting-started/
description: Learn how to install the Operator with Helm charts
title: Install the Operator with Helm
weight: 100
---
# Install the Operator with Helm

In this guide, you'll learn how to deploy [Grafana Agent Operator]({{< relref "./_index.md" >}}) into your Kubernetes cluster using the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator). To learn how to deploy Agent Operator without using Helm, see [Install Grafana Agent Operator]({{< relref "./getting-started.md" >}}).

> **Note**: If you are shipping your data to Grafana Cloud, use [Kubernetes Monitoring](/docs/grafana-cloud/kubernetes-monitoring/) to set up Agent Operator. Kubernetes Monitoring provides a simplified approach and preconfigured dashboards and alerts.

## Before you begin

To deploy Agent Operator with Helm, make sure that you have the following:

- A Kubernetes cluster
- The [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl) command-line client installed and configured on your machine
- The [`helm`](https://helm.sh/docs/intro/install/) command-line client installed and configured on your machine

> **Note:** Agent Operator is currently in beta and its custom resources are subject to change.

## Install the Agent Operator Helm Chart

In this section, you'll install the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator) into your Kubernetes cluster. This will install the latest version of Agent Operator and its [Custom Resource Definitions](https://github.com/grafana/agent/tree/main/operations/agent-static-operator/crds) (CRDs). The chart configures Operator to maintain a Service that lets you scrape kubelets using a `ServiceMonitor`.

To install the Agent Operator Helm chart:

1. Add and update the `grafana` Helm chart repo:

    ```bash
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    ```

1. Install the chart, replacing `my-release` with your release name:

    ```bash
    helm install my-release grafana/grafana-agent-operator
    ```

    If you want to modify the default parameters, you can create a `values.yaml` file and pass it to `helm install`:

    ```bash
    helm install my-release grafana/grafana-agent-operator -f values.yaml
    ```

    If you want to deploy Agent Operator into a namespace other than `default`, use the `-n` flag:

    ```bash
    helm install my-release grafana/grafana-agent-operator -f values.yaml -n my-namespace
    ```
    You can find a list of configurable template parameters in the [Helm chart repository](https://github.com/grafana/helm-charts/blob/main/charts/agent-operator/values.yaml).

1. Once you've successfully deployed the Helm release, confirm that Agent Operator is up and running:

    ```bash
    kubectl get pod
    kubectl get svc
    ```

    You should see an Agent Operator Pod in `RUNNING` state, and a `kubelet` service. Depending on your setup, this could take a moment.

## Deploy the Grafana Agent Operator resources

 Agent Operator is now up and running. Next, you need to install a Grafana Agent for Agent Operator to run for you. To do so, follow the instructions in the [Deploy the Grafana Agent Operator resources]({{< relref "./deploy-agent-operator-resources.md" >}}) topic. To learn more about the custom resources Agent Operator provides and their hierarchy, see [Grafana Agent Operator architecture]({{< relref "./architecture" >}}).
