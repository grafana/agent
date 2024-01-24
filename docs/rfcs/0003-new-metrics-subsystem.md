# New metrics subsystem

* Date: 2021-11-29
* Author: Robert Fratto (@rfratto)
* PR: [grafana/agent#1140](https://github.com/grafana/agent/pull/1140)
* Status: Abandoned

## Background

There are several open issues discussing major changes to the metrics
subsystem:

* [#872][#872]: Per-target sharding
* [#873][#873]: Reduce operational modes
* [#875][#875]: Introduce agent-wide clustering mechanism
* [#888][#888]: Remove internal instance manager system

These are significant changes to the code base. With the exception of #872, all
of the changes are mainly to reduce technical debt. The implementation effort
and lack of end-user benefits make them hard to schedule, despite being
genuinely beneficial for the maintenance of the project.

This proposal suggests a redesign of the metrics subsystem which has native
support for target sharding and lacks the technical debt from the current
subsystem.

## Goals

* Enable dynamically target scraping with:
  * Automatic scaling
  * Automatic failover
  * Target distribution

## Non-Goals

* Interaction with this new subsystem from existing subsystems
* Utilization of the configuration management API

## Implementation

Given the size of the change, work on the new subsystem should be done in a new
package (e.g., `pkg/metrics/next`), and exposed as an experimental change
hidden behind a feature flag (e.g., `--enable-features=metrics-next`).

## Design

The existing metrics subsystem is focused around a runtime-configurable Metrics
Instance system. Metrics Instances are primarily sourced from the config file
(through the `metrics.configs` array), but can also be dynamically added when
using integrations or the scraping service.

The new metrics subsystem break Metrics Instances up into multiple co-operating
parts:

1. [Discoverers](#Discoverers)
2. [Scrapers](#Scrapers)
3. [Senders](#Senders)

A Metrics Instance still exists conceptually, and is configured as normal
through the `metrics.configs` array. However, there will no longer be an
internal CRUD interface for dynamically managing them.

Finally, an agent-wide clustering mechanism will be added. This clustering
mechanism will allow agents to be aware of other running agents, and will
expose methods for an individual agent to determine ownership of a resource.
The [Clustering](#Clustering) section will describe how this works in detail.

All agents in the cluster will implement Discovers, Scrapers, and Senders.

```
                         +------------+ +------------+
Scrape Configs           | Config  A  | |  Config  B |
                         +------------+ +------------+
                                \              /
(1) SD Distribution              +------------/-------------+
                            v----------------+               \
                         +------------+ +------------+ +------------+
(2) Discoverers          |  Agent  A  | |  Agent  B  | |  Agent  C  |
                         +------------+ +------------+ +------------+
                             \    \                         /
(3) Target Distribution       +----+---------+-------------/---+
                            v-----------------\-----------+     \
                         +------------+ +------------+ +------------+
(4) Scrapers & Senders   |  Agent  A  | |  Agent  B  | |  Agent  C  |
                         +------------+ +------------+ +------------+

+============================================================+
||                                                          ||
|| (1) scrape_configs from runtime config are distributed   ||
||     amongst agents. Agent A owns Config B. Agent C owns  ||
||     Config A.                                            ||
||                                                          ||
|| (2) Agents perform service discovery for scrape configs. ||
||                                                          ||
|| (3) Agents partition discovered targets amongst cluster. ||
||     Agent A finds targets for Agent B and C. Agent C     ||
||     finds targets for Agent A.                           ||
||                                                          ||
|| (3) Agents partition discovered targets amongst cluster. ||
|| (4) Agents scrape targets from partitions they were sent ||
||     and write metrics to WAL which is picked up by       ||
||     remote_write.                                        ||
||                                                          ||
+===========================================================+
```

### Discoverers

Discoverers discover Prometheus targets and distribute them to Scrapers across
the cluster. There is one Discoverer per Metrics Instance in the
`metrics.configs` array from the agent's runtime config.

Each Discoverer runs a single Prometheus SD manager. The Discoverer will be
launched only with the set of SD jobs that the local agent owns, using the job
name as the ownership key. This introduces one layer of sharding, where each SD
job will only have one agent responsible for it. Note that relabeling rules are
not applied by the Discoverer.

Discovered targets are flushed to Scrapers in multiple partitions. Partitions
contain a set of targets owned by the same agent in the cluster, and introduces
the second (and last) layer of sharding, where each target will only have one
agent responsible for it. Partitions also include the Metrics Instance name,
since the same job may exist across multiple instances. The `__address__` label
from the target is used as the ownership key. Once all partitions are created,
they are sent to the corresponding agents over gRPC. Partitions that are owned
by the same agent as the Discoverer may be sent through a non-network
mechanism.

A partition will be created and sent to all agents in the cluster, even if the
partition is empty. This allows agents to know when they can stop scraping
something from a previous received partition.

Discovered targets will be re-flushed whenever the set of agents in the cluster
changes.

### Scrapers

Scrapers receive Prometheus targets from a Discoverer and scrape them,
appending scraped metrics to a Sender.

Specifically, Scrapers manage a dynamic set of Prometheus scrape managers. One
scrape manager will exist per instance that has a non-empty target partition.
Scrape managers will then be configured with the scrape jobs (including
relabeling rules) if they received at least one target for that job. The
definition of a scrape job is retrieved using the agent's runtime config.

There may be more than one Discoverer performing SD. This means that a Scraper
can expect to receive target partition from multiple Discoverers, and that it
needs a way to merge those partitions to determine the full set of targets to
scrape.

Scrapers utilize the knowledge that each targets from a scrape job are owned by
exactly one Discoverer. This allows the merge logic to be simple: store targets
by scrape job name which can be flattened into a single set. Jobs that do not
exist in the agent's runtime config will be ignored when merging, and
eventually removed in the background to limit memory growth.

With a set of targets, Scrapers will perform relabeling rules, scrape targets,
perform metric relabeling rules, and finally send the metrics to a Sender that
is associated with the Instance name from the partition.

### Senders

Finally, Senders store data in a WAL and configure Prometheus remote_write to
ship the WAL metrics to some remote system.

There is one sender launched per Metrics Instance from the agent configuration
file. Because other subsystems append samples to the WAL for delivery, Senders
must always exist, even if there aren't any Scrapers sending metrics to them.

The set of running Senders and their individual configurations will update
whenever the agent's configuration file changes.

### Clustering

An agent-wide cluster is always available, even if the local agent is not
connected to any remote agents.

The cluster will initially use [grafana/ckit][ckit], an extremely light
clustering toolkit that uses gossip for peer discovery and health checking. A
hash ring is locally deterministically calculated based on known peers.

Normally, gossip is done over a dedicated UDP connection to transmit messages
between peers. Since gossip is only utilized here for the peer list and health
checking, gossip is done over the existing gRPC protocol. This has the added
benefits for health checking the gRPC connection directly and reducing the
amount of things to configure when setting up clustering.

Bootstrapping the cluster will be done through [go-discover][go-discover] and a
`--cluster.discover-peers` command-line flag. This flag will be required to use
clustering, otherwise agents will act as a one-node cluster.

## Changes from the original design

### No partition TTL

The [original proposal][per-target sharding] for target-level sharding used a
TTL to detect if targets from jobs have gone stale. This added unnecessary
complexity to the implementation, and introduced bugs where clock drift could
cause targets to go stale immediately.

This new design avoids the need for a TTL by instead checking to see if an
entire job has gone stale using the runtime configuration.

## Edge Cases

### Discoverer network partition

A Discoverer network partition occurs when two Discoverers determine ownership
of the same job. This will cause targets to be sent twice to Scrapers. If
targets are sent to the same Scraper, no negative effect will occur: the
merging logic of scrapers will ignore the first partition and use the second
instead.

However, if targets are sent to different scrapers, then a Scraper network
partition occurs. This may also cause some targets to not be scraped by any
agent, depending on the order in which partitions are received by Discoverers.
Future changes may add resistance to ordering problems by using Lamport clocks.

### Scraper network partition

If two Scrapers are scraping the same target, Remote Write will reject the
duplicate samples. Otherwise, no noticeable effect occurs.

### Unhealthy Discoverer

Targets sent by the unhealthy Discoverer will continue to be active. Once the
unhealthy Discoverer is removed from the gossip memberlist, a new Discoverer
will pick up its SD jobs and re-deliver targets to the appropriate Scrapers.

### Unhealthy Scraper

Targets owned by the Scraper will be unscraped for a brief period of time. The
Scraper will be removed from the gossip memberlist, and force Discoverers to
re-flush targets. The targets will then be assigned to a new Scraper and the
system state will recover.

### Cluster networking failure

Nodes must be required to communicate with one another. If this is not
possible, the gossip memberlist will remove unreachable nodes and cause one or
more network partitions.

## Trade-offs

### No runtime instance management

This approach removes runtime instance management by using the loaded
configuration file as the source of truth. Subsystems that previously
dynamically launched instances can work around this by mutating the runtime
config when the config is first loaded.

### Complexity

Using the network for distribution adds some level of complexity and fragility
to the system. There may be unidentified edge cases or flaws in the designed
proposed here.

### No Configuration Store API

This approach doesn't support an external configuration store API. Such an API
should be delegated to an external process that flushes state to a file for the
agent to read.

### Configuration Desync

This approach requires all agents have the same configuration file. This can be
worked around by using [#1121][#1121] to help make sure all agents pull their
configs from the same source. A new metric that hashes the runtime config can
also enable alerting on config desync.

[#872]: https://github.com/grafana/agent/issues/872
[#873]: https://github.com/grafana/agent/issues/873
[#875]: https://github.com/grafana/agent/issues/875
[#888]: https://github.com/grafana/agent/issues/888
[per-target sharding]: https://docs.google.com/document/d/1JI804iaut6bKvZprOydes3Gb5Awo_J0sX-3ORlyc5l0
[ckit]: https://github.com/grafana/ckit
[go-discover]: https://github.com/hashicorp/go-discover
[#1121]: https://github.com/grafana/agent/issues/1121
