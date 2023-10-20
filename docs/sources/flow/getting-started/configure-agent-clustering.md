---
aliases:
- /docs/grafana-cloud/agent/flow/getting-started/configure-agent-clustering/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/configure-agent-clustering/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/getting-started/configure-agent-clustering/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/configure-agent-clustering/
menuTitle: Configure Grafana Agent clustering
title: Configure Grafana Agent clustering in an existing installation
description: Learn how to configure Grafana Agent clustering in an existing installation
weight: 400
---

# Configure Grafana Agent clustering in an existing installation

You can configure Grafana Agent to run with [clustering][] so that
individual agents can work together for workload distribution and high
availability.

{{% admonition type="note" %}}
Clustering is a [beta](http://www.grafana.com/docs/agent/flow/stability.md#beta) feature.
Beta features are subject to breaking changes and may be replaced with equivalent functionality that covers the same use case.
{{%/admonition %}}

This topic describes how to add clustering to an existing installation.

## Configure Grafana Agent clustering with Helm Chart

This section will guide you through enabling clustering when Grafana Agent is
installed on Kubernetes using the [Grafana Agent Helm chart][install-helm].

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

{{% docs/reference %}}
[clustering]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/clustering.md"
[clustering]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/flow/concepts/clustering.md"
[install-helm]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/setup/install/kubernetes.md"
[install-helm]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/kubernetes.md"
[UI]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging.md#component-detail-page"
[UI]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/flow/monitoring/debugging.md#component-detail-page"
{{% /docs/reference %}}
