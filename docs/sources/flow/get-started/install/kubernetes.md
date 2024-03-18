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

## Before you begin

* Install [Helm][] on your computer.
* Configure a Kubernetes cluster that you can use for {{< param "PRODUCT_NAME" >}}.
* Configure your local Kubernetes context to point to the cluster.

## Deploy

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

1. Verify that the {{< param "PRODUCT_NAME" >}} pods are running:

   ```shell
   kubectl get pods --namespace <NAMESPACE>
   ```

   Replace the following:

   - _`<NAMESPACE>`_: The namespace used in the previous step.

You have now successfully deployed {{< param "PRODUCT_NAME" >}} on Kubernetes,
using default Helm settings. In order to configure {{< param "PRODUCT_NAME" >}},
see the [Configure {{< param "PRODUCT_NAME" >}} on Kubernetes][Configure] guide.

## Next steps

- [Configure {{< param "PRODUCT_NAME" >}} on Kubernetes][Configure]

- [Check out {{< param "PRODUCT_NAME" >}} Helm chart documentation on Artifact Hub][Artifact Hub]

[Artifact Hub]: https://artifacthub.io/packages/helm/grafana/grafana-agent

[Helm]: https://helm.sh

{{% docs/reference %}}
[Configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/configure/configure-kubernetes.md"
[Configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-kubernetes.md"
{{% /docs/reference %}}
