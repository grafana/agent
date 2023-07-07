---
title: Distribute Prometheus metrics scrape load with clustering
weight: 500
---


# Distribute Prometheus metrics scrape load with clustering

A good predictor for the size of a Grafana Agent Flow deployment is the number
of targets they're scraping. [Clustering][] with target auto-distribution
allows a fleet of agents to work together to dynamically distribute their
scrape load, providing high-availability.

## Before you begin

- Have a [Prometheus metrics collection][] pipeline in place.
- Have a [clustered] Grafana Agent Flow StatefulSet set up.

## Steps

To distribute Prometheus metrics scrape load with clustering:

1. Locate your Grafana Agent Flow configuration file.
1. Within a Prometheus metrics collection pipeline, locate the
   `prometheus.scrape` component that will opt-in to auto-distributing its
   targets within the cluster.
1. Ensure that all `prometheus.scrape` component instances within the cluster
   have the same input target set in their Arguments.
1. Paste the following inside of the `prometheus.scrape` component arguments block
    ```
    clustering {
      enabled = true
    }
    ```
1. Reload the configuration using the `/-/reload` endpoint.
1. Use the [Flow UI] to check the details page of the `prometheus.scrape`
   component and its Debug Info section to verify that the targets are being
distributed between cluster peers.

[Clustering]: {{< relref "../concepts/clustering.md" >}}
[Prometheus metrics collection]: {{< relref "collect-prometheus-metrics.md" >}}
[clustered]: {{< relref "deploy-agent-cluster.md" >}}
[Flow UI]: {{< relref "../monitoring/debugging.md#component-detail-page" >}}
