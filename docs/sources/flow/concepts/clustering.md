---
aliases:
- /docs/grafana-cloud/agent/flow/concepts/clustering/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/concepts/clustering/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/concepts/clustering/
canonical: https://grafana.com/docs/agent/latest/flow/concepts/clustering/
labels:
  stage: beta
menuTitle: Clustering
title: Clustering (beta)
description: Learn about Grafana Agent clustering concepts
weight: 500
refs:
  prometheus.operator.podmonitors:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.operator.podmonitors/#clustering-beta
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.operator.podmonitors/#clustering-beta
  run:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/cli/run/#clustering-beta
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/cli/run/#clustering-beta
  clustering-page:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/#clustering-page
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/#clustering-page
  pyroscope.scrape:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.scrape/#clustering-beta
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/pyroscope.scrape/#clustering-beta
  debugging:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/#debugging-clustering-issues
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/#debugging-clustering-issues
  prometheus.scrape:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.scrape/#clustering-beta
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.scrape/#clustering-beta
  prometheus.operator.servicemonitors:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.operator.servicemonitors/#clustering-beta
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.operator.servicemonitors/#clustering-beta
---

# Clustering (beta)

Clustering enables a fleet of agents to work together for workload distribution
and high availability. It helps create horizontally scalable deployments with
minimal resource and operational overhead.

To achieve this, Grafana Agent makes use of an eventually consistent model that
assumes all participating Agents are interchangeable and converge on using the
same configuration file.

The behavior of a standalone, non-clustered agent is the same as if it was a
single-node cluster.

You configure clustering by passing `cluster` command-line flags to the [run](ref:run)
command.

## Use cases

### Target auto-distribution

Target auto-distribution is the most basic use case of clustering; it allows
scraping components running on all peers to distribute scrape load between
themselves. For target auto-distribution to work correctly, all agents in the
same cluster must be able to reach the same service discovery APIs and must be
able to scrape the same targets.

You must explicitly enable target auto-distribution on components by defining a
`clustering` block, such as:

```river
prometheus.scrape "default" {
    clustering {
        enabled = true
    }

    ...
}
```

A cluster state change is detected when a new node joins or an existing node goes away. All participating components locally
recalculate target ownership and rebalance the number of targets they’re
scraping without explicitly communicating ownership over the network.

Target auto-distribution allows you to dynamically scale the number of agents to distribute workload during peaks.
It also provides resiliency because targets are automatically picked up by one of the node peers if a node goes away.

Grafana Agent uses a fully-local consistent hashing algorithm to distribute
targets, meaning that, on average, only ~1/N of the targets are redistributed.

Refer to component reference documentation to discover whether it supports
clustering, such as:

- [prometheus.scrape](ref:prometheus.scrape)
- [pyroscope.scrape](ref:pyroscope.scrape)
- [prometheus.operator.podmonitors](ref:prometheus.operator.podmonitors)
- [prometheus.operator.servicemonitors](ref:prometheus.operator.servicemonitors)

## Cluster monitoring and troubleshooting

To monitor your cluster status, you can check the Flow UI [clustering page](ref:clustering-page).
The [debugging](ref:debugging) topic contains some clues to help pin down probable
clustering issues.

