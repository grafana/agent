---
title: Configure agent clustering with Helm
weight: 400
---

# Configure agent clustering with Helm

Grafana Agent Flow can be configured to run with [clustering][] so that
individual agents can work together for workload distribution and high
availability.

{{% admonition type="note" %}}
Clustering is a [beta][] feature. Beta features are subject to breaking
changes, and may be replaced with equivalent functionality that cover the same
use case.
{{%/admonition %}}

This topic describes how to add clustering to an existing installation.

[clustering]: {{< relref "../concepts/clustering.md" >}}
[beta]: {{< relref "../../stability.md#beta" >}}

## Before you begin

- [Install][] flow mode on Kubernetes using Helm.
    - Ensure that your `values.yaml` file sets `controller.type` to
      `statefulset`.

[Install]: {{< relref "../setup/install/kubernetes.md" >}}

## Steps

To configure clustering:

1. Amend your existing values.yaml file to add `clustering.enabled=true` inside
   of the `agent` block:

   ```yaml
   agent:
     clustering:
       enabled: true
   ```

1. Upgrade your installation to use the new values.yaml file:

   ```bash
   helm upgrade RELEASE_NAME -f values.yaml
   ```

   Replace `RELEASE_NAME` with the name of the installation you chose when you
   installed the Helm chart.

1. Use [UI][] to verify the cluster status:

   1. Click **Clustering** in the navigation bar.

   2. Ensure that all expected nodes appear in the resulting table.

[UI]: {{< relref "../monitoring/debugging.md#clustering-page" >}}
