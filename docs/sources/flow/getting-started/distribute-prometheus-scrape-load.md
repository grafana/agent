---
aliases:
- /docs/grafana-cloud/agent/flow/getting-started/distribute-prometheus-scrape-load/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/distribute-prometheus-scrape-load/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/getting-started/distribute-prometheus-scrape-load/
- /docs/grafana-cloud/send-data/agent/flow/getting-started/distribute-prometheus-scrape-load/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/distribute-prometheus-scrape-load/
description: Learn how to distribute your Prometheus metrics scrape load
menuTitle: Distribute Prometheus metrics scrape load
title: Distribute Prometheus metrics scrape load
weight: 500
---

# Distribute Prometheus metrics scrape load

A good predictor for the size of an agent deployment is the number of
Prometheus targets each agent scrapes. [Clustering][] with target
auto-distribution allows a fleet of agents to work together to dynamically
distribute their scrape load, providing high-availability.

> **Note:** Clustering is a [beta][] feature. Beta features are subject to breaking
> changes and may be replaced with equivalent functionality that covers the same use case.

## Before you begin

- Familiarize yourself with how to [configure existing Grafana Agent installations][configure-grafana-agent].
- [Configure Prometheus metrics collection][].
- [Configure clustering][] of agents.
- Ensure that all of your clustered agents have the same configuration file.

## Steps

To distribute Prometheus metrics scrape load with clustering:

1. Add the following block to all `prometheus.scrape` components which
   should use auto-distribution:

   ```river
   clustering {
     enabled = true
   }
   ```

2. Restart or reload agents for them to use the new configuration.

3. Validate that auto-distribution is functioning:

   1. Using the [Grafana Agent UI][UI] on each agent, navigate to the details page for one of
      the `prometheus.scrape` components you modified.

   2. Compare the Debug Info sections between two different agents to ensure
      that they're not scraping the same sets of targets.

{{% docs/reference %}}
[Clustering]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/clustering.md"
[Clustering]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/concepts/clustering.md"
[beta]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/stability.md#beta"
[beta]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/stability.md#beta"
[configure-grafana-agent]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/setup/configure"
[configure-grafana-agent]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/setup/configure"
[Configure Prometheus metrics collection]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/collect-prometheus-metrics.md"
[Configure Prometheus metrics collection]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/getting-started/collect-prometheus-metrics.md"
[Configure clustering]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/configure-agent-clustering.md"
[Configure clustering]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/getting-started/configure-agent-clustering.md"
[UI]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging.md#component-detail-page"
[UI]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging.md#component-detail-page"
{{% /docs/reference %}}
