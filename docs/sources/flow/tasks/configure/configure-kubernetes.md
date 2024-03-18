---
aliases:
- /docs/grafana-cloud/agent/flow/tasks/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/tasks/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/tasks/configure/configure-kubernetes/
- /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-kubernetes/
# Previous page aliases for backwards compatibility:
- /docs/grafana-cloud/agent/flow/setup/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/configure/configure-kubernetes/
- /docs/grafana-cloud/send-data/agent/flow/setup/configure/configure-kubernetes/
- ../../setup/configure/configure-kubernetes/ # /docs/agent/latest/flow/setup/configure/configure-kubernetes/
canonical: https://grafana.com/docs/agent/latest/flow/tasks/configure/configure-kubernetes/
description: Learn how to configure Grafana Agent Flow on Kubernetes
menuTitle: Kubernetes
title: Configure Grafana Agent Flow on Kubernetes
weight: 200
---

# Configure {{% param "PRODUCT_NAME" %}} on Kubernetes

This page describes how to apply a new configuration to {{< param "PRODUCT_NAME" >}}
when running on Kubernetes with the Helm chart. It assumes that:

- You have [installed {{< param "PRODUCT_NAME" >}} on Kubernetes using the Helm chart][k8s-install].
- You already have a new {{< param "PRODUCT_NAME" >}} configuration that you
  want to apply to your Helm chart installation.

If instead you're looking for help in configuring {{< param "PRODUCT_NAME" >}} to perform a specific task,
for example, to collect and store a specific telemetry signal, check out the [Kubernetes tasks section][k8s-tasks].

[k8s-tasks]: {{< relref "../kubernetes/" >}}
[k8s-install]: {{< relref "../../get-started/install/kubernetes" >}}

## Configure the Helm chart

To modify {{< param "PRODUCT_NAME" >}}'s Helm chart configuration, perform the following steps:

1. Create a local `values.yaml` file with a new Helm chart configuration.

   1. You can use your own copy of the values file or download a copy of the
      default [values.yaml][].

   1. Make changes to your `values.yaml` to customize settings for the
      Helm chart.

      Refer to the inline documentation in the default [values.yaml][] for more
      information about each option.

1. Run the following command in a terminal to upgrade your {{< param "PRODUCT_NAME" >}} installation:

   ```shell
   helm upgrade --namespace <NAMESPACE> <RELEASE_NAME> grafana/grafana-agent -f <VALUES_PATH>
   ```
   Replace the following:
   - _`<NAMESPACE>`_: The namespace you used for your {{< param "PRODUCT_NAME" >}} installation.
   - _`<RELEASE_NAME>`_: The name you used for your {{< param "PRODUCT_NAME" >}} installation.
   - _`<VALUES_PATH>`_: The path to your copy of `values.yaml` to use.

[values.yaml]: https://raw.githubusercontent.com/grafana/agent/main/operations/helm/charts/grafana-agent/values.yaml

### Kustomize considerations

If you are using [Kustomize][] to inflate and install the [Helm chart][], be careful
when using a `configMapGenerator` to generate the ConfigMap containing the
configuration. By default, the generator appends a hash to the name and patches
the resource mentioning it, triggering a rolling update.

This behavior is undesirable for {{< param "PRODUCT_NAME" >}} because the
startup time can be significant, for example, when your deployment has a large
metrics Write-Ahead Log. You can use the [Helm chart][] sidecar container to
watch the ConfigMap and trigger a dynamic reload.

The following is an example snippet of a `kustomization` that disables this behavior:

```yaml
configMapGenerator:
  - name: grafana-agent
    files:
      - config.river
    options:
      disableNameSuffixHash: true
```

## Configure the {{< param "PRODUCT_NAME" >}}

This section describes how to modify the {{< param "PRODUCT_NAME" >}}
configuration which is stored in a ConfigMap in the Kubernetes cluster. There
are two methods to perform this task.

### Method 1: Modify the configuration in the values.yaml file
Use this method if you prefer to embed your {{< param "PRODUCT_NAME" >}} configuration in the Helm chart's `values.yaml` file.

1. Modify the configuration file contents directly in the `values.yaml` file:

   ```yaml
   agent:
     configMap:
       content: |-
         // Write your Agent config here:
         logging {
           level = "info"
           format = "logfmt"
         }
   ```

1. Run the following command in a terminal to upgrade your {{< param "PRODUCT_NAME" >}} installation:

   ```shell
   helm upgrade --namespace <NAMESPACE> <RELEASE_NAME> grafana/grafana-agent -f <VALUES_PATH>
   ```
   Replace the following:
   - _`<NAMESPACE>`_: The namespace you used for your {{< param "PRODUCT_NAME" >}} installation.
   - _`<RELEASE_NAME>`_: The name you used for your {{< param "PRODUCT_NAME" >}} installation.
   - _`<VALUES_PATH>`_: The path to your copy of `values.yaml` to use.

### Method 2: Create a separate ConfigMap from a file
Use this method if you prefer to write your {{< param "PRODUCT_NAME" >}} configuration in a separate file.

1. Write your configuration to a file, for example, `config.river`.

   ```river
   // Write your Agent config here:
   logging {
     level = "info"
     format = "logfmt"
   }
   ```

1. Create a ConfigMap called `agent-config` from the above file:

   ```shell
   kubectl create configmap --namespace <NAMESPACE> agent-config "--from-file=config.river=./config.river"
   ```
   Replace the following:
   - _`<NAMESPACE>`_: The namespace you used for your {{< param "PRODUCT_NAME" >}} installation.

1. Modify Helm Chart's configuration in your `values.yaml` to use the existing ConfigMap:

   ```yaml
     agent:
     configMap:
       create: false
       name: agent-config
       key: config.river
   ```

1. Run the following command in a terminal to upgrade your {{< param "PRODUCT_NAME" >}} installation:

   ```shell
   helm upgrade --namespace <NAMESPACE> <RELEASE_NAME> grafana/grafana-agent -f <VALUES_PATH>
   ```
   Replace the following:
   - _`<NAMESPACE>`_: The namespace you used for your {{< param "PRODUCT_NAME" >}} installation.
   - _`<RELEASE_NAME>`_: The name you used for your {{< param "PRODUCT_NAME" >}} installation.
   - _`<VALUES_PATH>`_: The path to your copy of `values.yaml` to use.

[Helm chart]: https://github.com/grafana/agent/tree/main/operations/helm/charts/grafana-agent
[Kustomize]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/
