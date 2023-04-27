---
title: Installing Grafana Agent Operator with Helm
weight: 110
---
# Installing Grafana Agent Operator with Helm

In this guide you'll learn how to deploy the [Grafana Agent Operator]({{< relref "./_index.md" >}}) into your Kubernetes cluster using the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator).

> **Note:** Agent Operator is currently in beta and its custom resources are subject to change as the project evolves. It currently supports the metrics and logs subsystems of Grafana Agent. Integrations and traces support is coming soon.

By the end of this guide, you'll have deloyed Agent Operator into your cluster.

## Prerequisites

Before you begin, make sure that you have the following available to you:

- A Kubernetes cluster
- The `kubectl` command-line client installed and configured on your machine
- The `helm` command-line client installed and configured on your machine

## Install Agent Operator Helm Chart

In this step you'll install the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator) into your Kubernetes cluster. This will install the latest version of Agent Operator and its [Custom Resource Definitions](https://github.com/grafana/agent/tree/main/production/operator/crds) (CRDs). By default the chart will configure the operator to maintain a Service that allows you scrape kubelets using a `ServiceMonitor`.

Begin by adding and updating the `grafana` Helm chart repo:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

Next, install the chart:

```bash
helm install my-release grafana/grafana-agent-operator
```

Replace `my-release` with your desired release name.

If you want to modify the default parameters, you can create a `values.yaml` file and pass it in to `helm install`:

```bash
helm install my-release grafana/grafana-agent-operator -f values.yaml
```

A list of configurable template parameters can be found in the [Helm chart repository](https://github.com/grafana/helm-charts/blob/main/charts/agent-operator/values.yaml).

If you want to deploy Agent Operator into a namespace other than `default`, use the `-n` flag:

```bash
helm install my-release grafana/grafana-agent-operator -f values.yaml -n my-namespace
```

Once you've successfully deployed the Helm release, confirm that Agent Operator is up and running:

```bash
kubectl get pod
kubectl get svc
```

You should see an Agent Operator Pod in `RUNNING` state, and a `kubelet` Service.

## Conclusion

With Agent Operator up and running, you can move on to setting up a `GrafanaAgent` custom resource. This will discover `MetricsInstance` and `LogsInstance` custom resources and endow them with Pod attributes (like requests and limits) defined in the `GrafanaAgent` spec. To learn how to do this, please see [Custom Resource Quickstart]({{< relref "./custom-resource-quickstart.md" >}}).
