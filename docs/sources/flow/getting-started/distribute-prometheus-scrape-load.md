---
aliases:
- /docs/grafana-cloud/agent/flow/getting-started/distribute-prometheus-scrape-load/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/distribute-prometheus-scrape-load/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/getting-started/distribute-prometheus-scrape-load/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/distribute-prometheus-scrape-load/
menuTitle: Distribute Prometheus metrics scrape load
title: Distribute Prometheus metrics scrape load
description: Learn how to distribute your Prometheus metrics scrape load
weight: 500
refs:
  clustering:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/concepts/clustering/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/concepts/clustering/
  configure-grafana-agent:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/setup/configure/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/setup/configure/
  configure-prometheus-metrics-collection:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/getting-started/collect-prometheus-metrics/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/getting-started/collect-prometheus-metrics/
  beta:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/stability/#beta
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/stability/#beta
  configure-clustering:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/getting-started/configure-agent-clustering/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/getting-started/configure-agent-clustering/
  ui:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/#component-detail-page
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/#component-detail-page
---

# Distribute Prometheus metrics scrape load

A good predictor for the size of an agent deployment is the number of
Prometheus targets each agent scrapes. [Clustering](ref:clustering) with target
auto-distribution allows a fleet of agents to work together to dynamically
distribute their scrape load, providing high-availability.

> **Note:** Clustering is a [beta](ref:beta) feature. Beta features are subject to breaking
> changes and may be replaced with equivalent functionality that covers the same use case.

## Before you begin

- Familiarize yourself with how to [configure existing Grafana Agent installations](ref:configure-grafana-agent).
- [Configure Prometheus metrics collection](ref:configure-prometheus-metrics-collection).
- [Configure clustering](ref:configure-clustering) of agents.
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

   1. Using the [Grafana Agent UI](ref:ui) on each agent, navigate to the details page for one of
      the `prometheus.scrape` components you modified.

   2. Compare the Debug Info sections between two different agents to ensure
      that they're not scraping the same sets of targets.

