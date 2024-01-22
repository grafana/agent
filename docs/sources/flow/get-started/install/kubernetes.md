---
aliases:
- /docs/grafana-cloud/agent/flow/get-started/install/kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/get-started/install/kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/get-started/install/kubernetes/
- /docs/grafana-cloud/send-data/agent/flow/get-started/install/kubernetes/
# Previous docs aliases for backwards compatibility:
- ../../install/kubernetes/ # /docs/agent/latest/flow/install/kubernetes/
- /docs/grafana-cloud/agent/flow/setup/install/kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/install/kubernetes/
- /docs/grafana-cloud/send-data/agent/flow/setup/install/kubernetes/
- ../../setup/install/kubernetes/ # /docs/agent/latest/flow/setup/install/kubernetes/
canonical: https://grafana.com/docs/agent/latest/flow/get-started/install/kubernetes/
description: Learn how to deploy Grafana Agent Flow on Kubernetes
menuTitle: Kubernetes
title: Deploy Grafana Agent Flow on Kubernetes
weight: 200
---

# Deploy {{% param "PRODUCT_NAME" %}} on Kubernetes

You can deploy {{< param "PRODUCT_NAME" >}} on Kubernetes using our Helm chart.

{{< admonition type="note" >}}
These instructions show you how to install using our generic [Helm chart](https://github.com/grafana/agent/tree/main/operations/helm/charts/grafana-agent) for {{< param "PRODUCT_NAME" >}}.
You can deploy {{< param "PRODUCT_ROOT_NAME" >}} either in static mode or flow mode. The Helm chart deploys {{< param "PRODUCT_NAME" >}} by default.
{{< /admonition >}}

## Dedicated guides

We recommend using our dedicated guides for each telemetry signal to deploy a 
production-ready {{< param "PRODUCT_NAME" >}} on Kubernetes:

* For metrics, follow our dedicated [collect Prometheus metrics on Kubernetes][collect-prometheus] guide.
* For logs, follow our dedicated [collect logs on Kubernetes][collect-logs] guide.
* For anything else, follow the [generic Helm installation](#generic-helm-installation) instructions below.

[collect-prometheus]: {{< relref "../../tasks/kubernetes/collect-prometheus.md" >}}
[collect-logs]: {{< relref "../../tasks/kubernetes/collect-logs.md" >}}

If you want to collect multiple types of telemetry, we recommend deploying separate  
{{< param "PRODUCT_NAME" >}} workloads for each telemetry type.

## Generic Helm installation

Follow the instructions below to deploy {{< param "PRODUCT_NAME" >}} on Kubernetes using Helm.

### Before you begin

* Install [Helm][] on your computer.
* Configure a Kubernetes cluster that you can use for {{< param "PRODUCT_NAME" >}}.
* Configure your local Kubernetes context to point to the cluster.

### Deploy

To deploy {{< param "PRODUCT_ROOT_NAME" >}} on Kubernetes using Helm, run the following commands in a terminal window:

1. Add the Grafana Helm chart repository:

   ```shell
   helm repo add grafana https://grafana.github.io/helm-charts
   ```

1. Update the Grafana Helm chart repository:

   ```shell
   helm repo update
   ```
1. Create a namespace for the Agent:

   ```shell
   kubectl create namespace <NAMESPACE>
   ```

   Replace the following:

   - _`<NAMESPACE>`_: The namespace to use for your {{< param "PRODUCT_NAME" >}} 
     installation, such as `agent`.

1. Install {{< param "PRODUCT_ROOT_NAME" >}}:

   ```shell
   helm install --namespace <NAMESPACE> <RELEASE_NAME> grafana/grafana-agent
   ```

   Replace the following:
  - _`<NAMESPACE>`_: The namespace created in the previous step.
  - _`<RELEASE_NAME>`_: The name to use for your {{< param "PRODUCT_ROOT_NAME" >}} installation, such as `grafana-agent-flow`.

For more information on the {{< param "PRODUCT_ROOT_NAME" >}} Helm chart, refer to the Helm chart documentation on [Artifact Hub][].

[Artifact Hub]: https://artifacthub.io/packages/helm/grafana/grafana-agent

### Next steps

- [Configure {{< param "PRODUCT_NAME" >}}][Configure]

[Helm]: https://helm.sh

{{% docs/reference %}}
[Configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/configure/configure-kubernetes.md"
[Configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-kubernetes.md"
{{% /docs/reference %}}
