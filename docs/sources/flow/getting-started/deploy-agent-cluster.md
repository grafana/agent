---
title: Deploy agent clustering with Helm
weight: 400
---

# Deploy agent clustering with Helm

Grafana Agent Flow can be configured to run with [clustering][] so that
individual agents can work together for workload distribution and high
availability.

This topic describes how to use the Grafana Agent [Helm chart][] to deploy a
set of cluster-aware agents.


## Before you begin

- Install [Helm][]
- Ensure you can connect to your Kubernetes cluster.
- Set up the Grafana chart repository. 
    ```
    $ helm repo add grafana https://grafana.github.io/helm-charts
    $ helm repo update
    ```

## Steps

To deploy agent clustering with Helm:

1. Create a new values.yaml file, or amend a current one with the following
   block
    ```yaml
    agent:
      clustering:
        enabled: true
    controller:
      type: 'statefulset'
    ```

1. Use `helm install` to install the grafana/grafana-agent Helm chart on your
   Kubernetes cluster. Replace `RELEASE_NAME` with the desired name for the
   installation.
    ```
    $ helm install RELEASE_NAME grafana/grafana-agent -f values.yaml
    ```

1. Use the [Flow UI] page to verify the cluster status.

1. Use `helm upgrade` to upgrade the `RELEASE_NAME` installation with new
   values.
    ```
    $ helm upgrade RELEASE_NAME -f values.yaml
    ```

[clustering]: {{< relref "../concepts/clustering.md" >}}
[Helm chart]: https://artifacthub.io/packages/helm/grafana/grafana-agent
[Helm]: https://helm.sh/
[Flow UI]: {{< relref "../monitoring/debugging.md#clustering-page" >}}
