# Agent Clustering

* Date: 2023-03-02
* Author: Paschalis Tsilias (@tpaschalis)
* PR: [grafana/agent#3151](https://github.com/grafana/agent/pull/3151)

## Summary - Background
We routinely run agents with 1-10 million active series; we regularly see
sharded agent deployments with ~30M series each run without hiccups.

Our usual recommendation is to start thinking about horizontal scaling around
the 2M mark. Unfortunately the current
[options](https://grafana.com/docs/agent/latest/operation-guide/) have a number
of challenges, and many of these are not even directly applicable to Flow mode:

* Hashmod sharding is not dynamic and requires _all_ agents to update their
  configuration and reshard whenever a member joins or leaves the cluster.
* The scraping service requires another set of dependencies to be introduced
  (etcd or consul), and can only shard on the configuration-file level which
puts the responsibility on the developer to maintain multiple configuration
files that ideally, also balance the number of targets they expose.
* Host filtering ties users to a daemonset-like deployment model and can
  unnecessary load on service discovery APIs.
* Hand-writing configuration to distribute the load into different agent
  deployments is simply not manageable.

As Flow mode aims to solve many of the configuration woes of static mode, we
would like to propose a new Flow-native clustering mode that allows the Agent
scale elastically with a single configuration file and an eventually consistent
model.

## Goals
* Implement a clustered mode that allows the Agent to elastically scale without
  changing the configuration
* Enable Flow components to work together and distribute load across a cluster
* Enable fine-grained scheduling of Flow components within a cluster
* Provide an easy-to-use replacement for scraping service and hashmod sharding
* Allow users to understand and debug the status of their cluster

## Non-goals
* Recreate the scraping service as-is. More specifically:
  - Use an external store for configuration files
  - Expose an API for managing configuration
  - Running multiple configuration files at once.
* Distribute load by merging multiple configuration files.

## Proposal
The proposal is based on prior art: https://github.com/grafana/agent/issues/872, https://github.com/grafana/agent/pull/1140

* We will continue with a
  [gossip-based](https://en.wikipedia.org/wiki/Gossip_protocol) approach using
Hashicorp’s memberlist for our cluster
* We will reuse the rfratto/ckit package code
* We will use HTTP2 for communication between nodes
* We will use go-discover for bootstrapping the cluster and discovering peers
* A non-clustered Agent will work similar to a one-node cluster, which in the
  future will be the default mode of operation

## Implementation
The feature will be behind a feature flag `--enable-features=clustering`. An
agent can opt-in to clustering by passing a `--cluster.discover-peers`
command-line flag with a comma-separated list of peers to connect to. Whenever
an agent node receives a message about another node joining or leaving the
cluster, it will propagate the message to its neighbors, and so on, until this
information has reached all members of the cluster. The gossip memberlist will
be utilized for the peer list, health checking and distribution of tokens
between the agent nodes.

All nodes will have access to a shared ckit.Sharder interface implementation
which will be used to expose methods for each individual agent’s Flow
controller to determine ownership of resources. As nodes enter and exit the
cluster, the Sharder (eg. a consistent hashing ring) will redistribute tokens
among the nodes in the cluster. The eventually consistent cluster state is when
all nodes are working with the same configuration and have knowledge of each
one of their peers in the cluster.The Sharder will be used to determine
ownership of resources by hashing a desired value and checking the peers
responsible for the corresponding ring token.

When all nodes in the cluster have an up-to-date image of their peers, they
will be able to independently agree to the ownership of a resource without
having to communicate with each other, as local hashing will provide the same
results for all nodes. For example, an agent node will be able to hash the
label set of a target and check which of the peers is responsible for scraping
that target. Similarly, an agent node will be able to hash the fully qualified
component name and decide whether a component needs to be scheduled on this
node or if another peer takes responsibility for it.

On a more practical note, this clustering will most likely work with a
[Kubernetes HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
so we can dynamically scale a fleet of agents. The target resource the
autoscaler will look out for is most likely memory usage. In the future we may
allow scaling on different/custom metrics as well.

We will start by creating the abstractions that will enable the two use-cases
presented in the next section.

## Use cases
In the first iteration of agent clustering, we would like to start with the
following use-cases. These two are distinct in the way that they make use of
scheduling.

The first one makes sure that we have a way of notifying components of cluster
changes and calling their Update method and continuously re-evaluate ownership
of resources. When this is implemented, we can start thinking about the second
one that would provide a way to dynamically enable and disable components in
the cluster.

### Distributing scrape targets
The main predictor for the size of an Agent deployment is the number of targets
it is scraping/reading logs from. Components that use the Flow concept of a
“target” as their Arguments should be able to distribute the target load
between themselves. To do that we can introduce a layer of abstraction over
the Targets definition that can interact with the Sharder provided by the
clusterer and provide a simple API, for example:
```go
type Targets interface {
    Get() []Target
}
```

When clustering is enabled, this API will allow the component to distribute the
targets using a consistent hashing approach. Ownership is determined using the
entire label set’s fingerprint/hash and where it belongs on the implementation
exposed by the shared Sharder interface and each agent node will only scrape
the targets that it is personally responsible for. When clustering is
disabled, the components should work as if they are a standalone node and
scrape the entire target set.

We will be using each component’s Arguments to enable/disable the clustering
functionality. The Targets abstraction will be registering and updating
metrics that will detail the distribution of the targets between nodes, as well
as how they change over time.

I propose that we start with the following set of components that make use of
this functionality: prometheus.scrape, loki.source.file,
loki.source.kubernetes, and pyroscope.scrape.

Here’s how the configuration for a component could look like:
```river
prometheus.scrape "pods" {
    clustering {
        node_updates = true
    }

    targets    = discovery.kubernetes.pods.targets
    forward_to = [prometheus.remote_write.default.receiver]
}
```

Distribution of load should happen automatically for these components without
any manual handling from the component authors. Components that enable
clustering will register themselves to the Flow controller to be notified and
have their Update method called whenever the state of the cluster changes. An
idea of how this can work is to have something similar to how OnExportsChange
works. This abstraction will allow components to communicate back to the
controller whether a clustering-related Argument has changed. The controller
will keep a record of these components and when a cluster change is detected
these will have their Update method called (or just queued for re-evaluation).

While the cluster is not in a consistent state, this might lead to temporarily
missing or duplicated scrapes as not all nodes will have the same image of the
cluster and agree on the distribution of tokens. As long as the cluster can
stabilize within a time period comparable to the scraping interval, this should
not be an issue.

Finally, scaling the cluster up and down results in only 1/N targets’ ownership
being transferred.

### Component scheduling
The exposed clustering functionality can also allow
for fine-grained scheduling of specific Flow components. The widely used
example is that, in the context of an agent cluster, we would only need to run
a MySQL exporter once per MySQL database. I propose that we create a new
health status called “Disabled”. Graph nodes should provide good hints in the
UI in regards to which agent node the component was scheduled at and work
similarly to Exited nodes. Disabled components should be registered and built
by the controller and _not_ have their Run method called. Downstream
dependencies will get the evaluated values of the components exports, but an
Disabled dependency should not have any other side-effects for now. This may
warrant component changes. For example, the initial value of a LogsReceiver is
a channel which will block fanning out to other entries; this should be fixed.

The controller will have access to the Sharder implementation and node
ownership is determined by hashing the ComponentNode ID that we want to
schedule. Once all the components have been loaded in, the controller will
check if it should exclude any components from the slice of `runnables` that
will be passed to the Flow Scheduler.

The same logic should be applied during cluster changes. In case where a new
node might get ownership of the component, this loop will call Synchronize with
the right set of components so that either they Run, or their context is
terminated.

Finally, when clustering is enabled, each component will expose a set of
metrics that use labels to announce the node it has been scheduled on. If this
is a worry due cardinality issues, we can find another way of providing this
information.

On a more practical note, we’ll have to choose how components might use to
opt-in to the component scheduling.

For example, we could implement either:
* Implicitly adding a new Argument block that is implicitly present by default on
_all_ components:
```
--- cfg.river ---
prometheus.scrape “default” {
  clustering {
    node_scheduling = true
  }
  ...
}
```

* Having a new method exposed from the component’s Options to enable/disable
clustering and component authors can decide when/how to call it:
```
func (c *Component) Update(args component.Arguments) error {
    newArgs := args.(Arguments)
    ...
    c.opts.EnableClusterScheduling(newArgs.Clustering)
}
```

* Having top-level configuration block that handles which components are
scheduled:
```
--- cfg.river ---
cluster_scheduling {
  enable = [prometheus.scrape.default]
}
```

Since we cannot predict which of the growing list of components (up to 57
currently, >110 planned) will require clustered scheduling, we should try and
find a higher-level abstraction for it so it can be used by _any_ component.
As such, I propose that we go with the _first_ option. It might be a little
harder to implement, but is most in-line with the high-level abstraction that
we’re aiming for.

## Failure Modes

### Configuration partition
One of our axioms is that all agents in the same cluster run the same
configuration file, can reach the same discovery APIs and remote endpoints, and
have the same environment variable and network access. In case that agents
_cannot_ run the same configuration file (eg. due to different versions), or
that network issues prevent them from discovering or remote-writing correctly,
it will be hard to debug and understand where the problem lies.

At the very least, we should report what we can control, and that is the hash
of the configuration file. Ideally, as a new configuration file is being
applied to an agent cluster (eg. pods being rolled out), the state will
eventually be consistent. Is this enough, or should we limit clustering to
only nodes that have the same configuration hash?

### Networking failures
In case that a node is unreachable due to networking issues, it will be removed
from the gossip memberlist and cause one or more network partitions.

Also, in case that agent nodes lose connectivity with their cluster peers but
not to scrape targets or remote write, they will fall back to behaving as
single-node clusters leading them to overload themselves. We can recommend
setting some limits per agent to avoid this, or have alerts to detect multiple
single-node clusters running with the same config hash.

### Scrape targets network partition
If two agents are scraping the same target (unbeknownst to each other), the
cluster will incur some extra load, but remote write will reject the duplicate
sample (first one wins).

### Scheduling network partition
If two agents are scheduling the same component (unbeknownst to each other),
similarly the cluster will incur some extra load, but the first sample wins.

### Unhealthy node
In case a node goes unhealthy then both its targets and scheduled components
will end up not scraping any metrics for a period of time. When the node is
removed from the memberlist, then the component will be rescheduled on another
node, and its targets will be redistributed elsewhere. As long as the amount
of time required for the cluster state to recover is (how much?) smaller than
the typical scrape interval, then this behavior might not result in losing any
scrape cycles.

## Debugging
The clustering implementation must provide tools so that users can understand
how clustering works at the agent node level, as well as the component level.

On the _node level_, we can introduce a new tab on the Flow UI page which shows
the status of all nodes in the cluster and allows users to navigate to their UI
page. On the _component level_, the component’s debug info will contain
information regarding both the entire target set _and_ the targets the current
node is responsible for, as well as an indication of the work that other
components are doing, and provide a way to navigate to that node’s UI.

We will expose some clustering-specific internal metrics that provide a view of
the cluster status, such as the hash of the configuration file applied, the
load on each cluster, the tokens the cluster is responsible for, the timestamp
it was last updated, as well as a set of dashboard that can give this
information out at a glance.

## Questions - Concerns - Limitations

### Component Arguments naming
I’m not yet 100% sold on the name of the Arguments that components can use to
enable clustering, I’m open to suggestions. More specifically, I’m not sure if
they should be tied to the specific use-case that we’re trying to achieve, or
be more generic.

### Incurring load on SD API
Having the clustering happen on the targets layer means that an N sized cluster
will require putting N-times more load on the service discovery API. This is
true even today with N hashmod shards, but still might be something to look out
for if we go for larger cluster sizes.

### Receiver-based components
The clustering approach is mainly useful for distributing _internal_ load in
pull-based pipelines. Push-based pipelines can use different external load
distribution mechanisms such as a load balancer placed in front of replicas so
the clustering approach described here is most likely not applicable.

 ## Future roadmap/ideas
* Should we enable replication from the clustering implementation itself? Eg.
  allow targets to belong to _two_ nodes?
* Should we make the target distribution strategy configurable? (eg. determine
  ownership by first grouping into `__address__` or some other field)
* Should the new “Disabled” health status propagate through the graph? Eg. if a
  prometheus.scrape component only scrapes a prometheus.exporter.mysql
exporter, it should only get scheduled where its dependency is and not
elsewhere.
