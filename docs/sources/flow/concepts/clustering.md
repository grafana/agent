---
title: Clustering
weight: 500
labels:
  stage: beta
---

# Clustering (beta)

Clustering enables Grafana Agent Flow to coordinate a fleet of agents working
together for workload distribution and high availability. It helps create
horizontally scalable deployments with minimal resource and operational
overhead.

To achieve this, Grafana Agent makes use of an eventually consistent model that
assumes all participating Agents are interchangeable and converge on using the
same configuration file.

The behavior of a standalone, non-clustered agent is the same as if it was a
single-node cluster.

## Use cases

[Setting up][] clustering using the command-line arguments is the first step in
making agents aware of one another.

Components that support clustering can define the `clustering` block in their
River config and need explicitly opt-in to participating to a clustering use
case.

For example, the `prometheus.scrape` component can opt-in to auto-distributing
targets between nodes by adding the `clustering` block with the `enabled`
argument set to `true`:
```river
prometheus.scrape "default" {
    clustering {
      enabled = true
    }
    ...
}
```

### Target auto-distribution

Target auto-distribution is the most basic use case of clustering; it allows
scraping components running on all peers to distribute scrape load between
themselves. All nodes must have access to the same service discovery APIs, and
the set of targets should converge on a timeline comparable to the scrape
interval.

Whenever a cluster state change is detected, either due to a new node joining
or an existing node going away, all participating components locally
recalculate target ownership and rebalance the number of targets they’re
scraping without explicitly communicating ownership over the network.

As such, target auto-distribution not only allows to dynamically scale the
number of agents to distribute workload during peaks, but also provides
resiliency, since in the event of a node going away, its targets are
automatically picked up by one of their peers. 

The agent makes use of a fully-local consistent hashing algorithm to distribute
targets, meaning that on average only ~1/N of the targets are redistributed.

Here's some of the components that that support target auto-distribution: 
- [prometheus.scrape][]
- [pyroscope.scrape][]
- [prometheus.operator.podmonitors][]
- [prometheus.operator.servicemonitors][]

## Cluster monitoring and troubleshooting

To monitor your cluster status, you can check the Flow UI [clustering page][].
The [debugging][] topic contains some clues to help pin down probable
clustering issues.

[Setting up]: {{< relref "../reference/cli/run.md#clustering-beta" >}}
[clustering page]: {{< relref "../monitoring/debugging.md#clustering-page" >}}
[debugging]: {{< relref "../monitoring/debugging.md#debugging-clustering-issues" >}}

[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md#clustering-beta" >}}
[pyroscope.scrape]: {{< relref "../reference/components/pyroscope.scrape.md#clustering-beta" >}}
[prometheus.operator.podmonitors]: {{< relref "../reference/components/prometheus.operator.podmonitors.md#clustering-beta" >}}
[prometheus.operator.servicemonitors]: {{< relref "../reference/components/prometheus.operator.servicemonitors.md#clustering-beta" >}}
