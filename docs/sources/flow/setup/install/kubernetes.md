---
description: Learn how to install Grafana Agent Flow on Kubernetes
title: Install Grafana Agent Flow on Kubernetes
menuTitle: Kubernetes
weight: 100
aliases:
 - ../../install/kubernetes/
---

## Install Grafana Agent Flow on Kubernetes

Grafana Agent Flow can be installed on Kubernetes by using the Helm chart for
Grafana Agent Flow.

## Before you begin

* Ensure that you have [Helm][] installed on your system.
* Ensure that you have a Kubernetes cluster to deploy Grafana Agent Flow to.
* Ensure that your local Kubernetes context is configured to point to the
  correct cluster.

[Helm]: https://helm.sh

## Steps

> **NOTE**: These instructions install the generic [Helm chart][] for Grafana
> Agent, which can deploy Grafana Agent either in static mode or Flow mode.
> The Helm chart deploys Grafana Agent in Flow mode by default.
>
> [Helm chart]: https://github.com/grafana/agent/tree/main/operations/helm/charts/grafana-agent

To install Grafana Agent Flow on Kubernetes using Helm, perform the following
steps in a terminal:

1. Add the Grafana Helm chart repository:

   ```shell
   helm repo add grafana https://grafana.github.io/helm-charts
   ```

2. Ensure that the Grafana Helm chart repository is up to date:

   ```shell
   helm repo update
   ```

3. Install the Grafana Agent Helm chart:

   ```shell
   helm install RELEASE_NAME grafana/grafana-agent
   ```

   Replace `RELEASE_NAME` with a name to use for your Grafana Agent Flow
   installation, such as `grafana-agent-flow`.

For more information on the Grafana Agent Helm chart, refer to the Helm chart
on [Artifact Hub][].

[Artifact Hub]: https://artifacthub.io/packages/helm/grafana/grafana-agent

## Operation guide

### Customize deployment

To customize the deployment used to deploy Grafana Agent Flow on Kubernetes,
perform the following steps:

1. Download a local copy of [values.yaml][] for the Helm chart.

2. Make changes to your copy of `values.yaml` to customize settings for the
   Helm chart.

   Refer to inline documentation in the `values.yaml` to understand what each
   option does.

3. Run the following command in a terminal to upgrade your Grafana Agent Flow
   installation:

   ```shell
   helm upgrade RELEASE_NAME grafana/grafana-agent -f VALUES_PATH
   ```

   1. Replace `RELEASE_NAME` with the name you used for your Grafana Agent Flow
      installation.

   2. Replace `VALUES_PATH` with the path to your copy of `values.yaml` to use.

[values.yaml]: https://raw.githubusercontent.com/grafana/agent/main/operations/helm/charts/grafana-agent/values.yaml

### Kustomize considerations

If using [Kustomize][] to inflate and install the [Helm chart][], be careful
when using a `configMapGenerator` to generate the ConfigMap containing the
configuration. By default, the generator appends a hash to the name and patches
the resource mentioning it, triggering a rolling update.

In the case of Grafana Agent Flow, this behavior is undesirable, as the startup
time can be significant depending on the size of the Write-Ahead Log. Instead,
the [Helm chart][] provides a sidecar container that will watch the ConfigMap
and trigger a dynamic reload.

Here is an example snippet of a `kustomization` that disables this behavior:

```yaml
configMapGenerator:
  - name: grafana-agent
    files:
      - config.river
    options:
      disableNameSuffixHash: true
```

[Kustomize]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/
