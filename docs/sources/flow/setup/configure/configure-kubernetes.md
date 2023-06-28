---
description: Learn how to configure Grafana Agent on Kubernetes
title: Configure Grafana Agent on Kubernetes
menuTitle: Kubernetes
weight: 200
---

# Configure Grafana Agent on Kubernetes

## Customize deployment

To customize the deployment used to deploy Grafana Agent on Kubernetes,
perform the following steps:

1. Download a local copy of [values.yaml][] for the Helm chart.

2. Make changes to your copy of `values.yaml` to customize settings for the
   Helm chart.

   Refer to inline documentation in the `values.yaml` to understand what each
   option does.

3. Run the following command in a terminal to upgrade your Grafana Agent
   installation:

   ```shell
   helm upgrade RELEASE_NAME grafana/grafana-agent -f VALUES_PATH
   ```

   1. Replace `RELEASE_NAME` with the name you used for your Grafana Agent
      installation.

   2. Replace `VALUES_PATH` with the path to your copy of `values.yaml` to use.

[values.yaml]: https://raw.githubusercontent.com/grafana/agent/main/operations/helm/charts/grafana-agent/values.yaml

## Kustomize considerations

If using [Kustomize][] to inflate and install the [Helm chart][], be careful
when using a `configMapGenerator` to generate the ConfigMap containing the
configuration. By default, the generator appends a hash to the name and patches
the resource mentioning it, triggering a rolling update.

In the case of Grafana Agent, this behavior is undesirable, as the startup
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
