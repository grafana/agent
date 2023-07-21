---
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/configure-agent-clustering/
menuTitle: Configure Grafana Agent clustering
title: Configure Grafana Agent clustering in an existing installation
weight: 400
---

# Configure Grafana Agent clustering

You can configure Grafana Agent to run with [clustering][] so that
individual agents can work together for workload distribution and high
availability.

{{% admonition type="note" %}}
Clustering is a [beta][] feature. Beta features are subject to breaking
changes and may be replaced with equivalent functionality that covers the same
use case.
{{%/admonition %}}

This topic describes how to add clustering to an existing installation.

[clustering]: {{< relref "../concepts/clustering.md" >}}
[beta]: {{< relref "../../stability.md#beta" >}}

## Configure Grafana Agent clustering with Helm Chart

This section will guide you through enabling clustering when Grafana Agent is
installed on Kubernetes using the [Grafana Agent Helm chart][install-helm].

[install-helm]: {{< relref "../setup/install/kubernetes.md" >}}

### Before you begin

- Ensure that your `values.yaml` file has `controller.type` set to
  `statefulset`.

### Steps

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
