---
title: Distribute your Prometheus metrics scrape load
menuTitle: Distribute Prometheus metrics scrape load
weight: 500
---

# Distribute Prometheus metrics scrape load

A good predictor for the size of an agent deployment is the number of
Prometheus targets each agent scrapes. [Clustering][] with target
auto-distribution allows a fleet of agents to work together to dynamically
distribute their scrape load, providing high-availability.

{{% admonition type="note" %}}
Clustering is a [beta][] feature. Beta features are subject to breaking
changes and may be replaced with equivalent functionality that covers the same
use case.
{{%/admonition %}}

[beta]: {{< relref "../../stability.md#beta" >}}
[Clustering]: {{< relref "../concepts/clustering.md" >}}

## Before you begin

- Familiarize yourself with how to [configure existing installs][].
- [Configure Prometheus metrics collection][].
- [Configure clustering][] of agents.
- Ensure that all of your clustered agents have the same config file.

[configure existing installs]: {{< relref "../setup/configure" >}}
[Configure Prometheus metrics collection]: {{< relref "collect-prometheus-metrics.md" >}}
[Configure clustering]: {{< relref "./configure-agent-clustering.md" >}}

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

   1. Using the [UI][] on each agent, navigate to the details page for one of
      the `prometheus.scrape` components you modified.

   2. Compare the Debug Info sections between two different agents to ensure
      that they are not scraping the same sets of targets.

[UI]: {{< relref "../monitoring/debugging.md#component-detail-page" >}}
