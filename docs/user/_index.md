+++
title = "Grafana Agent documentation"
weight = 1
+++

# Grafana Agent

Grafana Agent is a telemetry collector for sending metrics, logs,
and trace data to the opinionated Grafana observability stack. It works best
with:

* [Grafana Cloud](https://grafana.com/products/cloud/)
* [Grafana Enterprise Stack](https://grafana.com/products/enterprise/)
* OSS deployments of [Grafana Loki](https://grafana.com/oss/loki/), [Prometheus](https://prometheus.io/), [Cortex](https://cortexmetrics.io/), and [Grafana Tempo](https://grafana.com/oss/tempo/)

The Agent supports collecting telemetry data by utilizing the same battle-tested
code from the official platforms. It uses Prometheus for metrics collection,
Grafana Loki for log collection, and [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector) for trace
collection.

Grafana Agent uses less memory on average than Prometheus – by doing less (only focusing on `remote_write`-related functionality).

Grafana Agent allows for deploying multiple instances of the Agent in a cluster and only scraping metrics from targets that running at the same host. This allows distributing memory requirements across the cluster rather than pressurizing a single node.

## Metrics

Unlike Prometheus, the Grafana Agent is _just_ targeting `remote_write`,
so some Prometheus features, such as querying, local storage, recording rules,
and alerts aren't present. `remote_write`, service discovery, and relabeling
rules are included.

The Grafana Agent has a concept of an "instance", each of which acts as
its own mini Prometheus agent with their own `scrape_configs` section and
`remote_write` rules. More than one instance is useful when you want to have
completely separate configs that write to two different locations without
needing to worry about advanced metric relabeling rules. Multiple instances also
come into play for the [Scraping Service Mode]({{< relref "./scraping-service" >}}).

The Grafana Agent can be deployed in three modes:

- Prometheus `remote_write` drop-in
- [Host Filtering mode](#host-filtering)
- [Scraping Service Mode]({{< relref "./scraping-service" >}})

The default deployment mode of the Grafana Agent is a _drop-in_
replacement for Prometheus `remote_write`. The Agent will act similarly to a
single-process Prometheus, doing service discovery, scraping, and remote
writing.

_Host Filtering mode_ is achieved by setting a `host_filter` flag on a specific
instance inside the Agent's configuration file. When this flag is set, the
instance will only scrape metrics from targets that are running on the same
machine as the instance itself. This is extremely useful to migrate to sharded
Prometheus instances in a Kubernetes cluster, where the Agent can be deployed as
a DaemonSet and distribute memory requirements across multiple nodes.

Note that Host Filtering mode and sharding your instances means that if an
Agent's metrics are being sent to an alerting system, alerts for that Agent may
not be able to be generated if the entire node has problems. This changes the
semantics of failure detection, and alerts would have to be configured to catch
agents not reporting in.

The final mode, _Scraping Service Mode_ 
clusters a subset of agents. It acts as a go-between for the drop-in mode
(which does no automatic sharding) and `host_filter` mode (which forces sharding
by node). The Scraping Service Mode clusters a set of agents with a set of
shared configs and distributes the scrape load automatically between them. For
more information, refer to ({{< relref "./scraping-service" >}}).

### Host filtering

Host filtering configures Agents to scrape targets that are running on the same
machine as the Grafana Agent process. It:

1. Gets the hostname of the agent by the `HOSTNAME` environment variable or
   through the default.
2. Checks if the hostname of the agent matches the label value for `__address__`
   service-discovery-specific node labels against the discovered target.

If the filter passes, the target is allowed to be scraped. Otherwise, the target
will be silently ignored and not scraped.

For detailed information on the host filtering mode, refer to the [operation
guide]({{< relref "./operation-guide#host-filtering-beta" >}}).

## Logs

Grafana Agent supports collecting logs and sending them to Loki using its
`loki` subsystem. This is done using the upstream
[Promtail](https://grafana.com/docs/loki/latest/clients/promtail/) client, which
is the official first-party log collection client created by the Loki
developer team.

## Traces

Grafana Agent supports collecting traces and sending them to Tempo using its
`traces` subsystem. This is done using the upstream [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).
Agent can ingest OpenTelemetry, OpenCensus, Jaeger, Zipkin, or Kafka spans.
See documentation on how to configure [receivers]({{< relref "./configuration/traces-config.md" >}}).
The agent is capable of exporting to any OpenTelemetry GRPC compatible system.

## Comparison to alternatives

Grafana Agent is optimized for [Grafana Cloud](https://grafana.com/products/cloud/),
but can be used while using an on-prem `remote_write`-compatible Prometheus API
and an on-prem Loki. Unlike alternatives, Grafana Agent extends the
official code with extra functionality. This allows the Agent to give an
experience closest to its official counterparts, unlike existing alternatives which
typically try to re-implement everything from scratch.

### Why not just use Telegraf?

Telegraf is a fantastic project and was actually considered as an alternative
to building our own agent.
It could work, but ultimately it was not chosen due to lacking service discovery
and metadata label propagation.
While these features could theoretically be added to Telegraf as OSS contributions,
there would be a lot of forced hacks involved due to its current design.

Additionally, Telegraf is a much larger project with its own goals for its community,
so any changes need to fit the general use cases it was designed for.

With the Grafana Agent as its own project, we can deliver a more curated agent
specifically designed to work seamlessly with Grafana Cloud and other
`remote_write` compatible Prometheus endpoints as well as Loki for logs
and Tempo for traces, all-in-one.

## Next steps

For more information on installing and running the agent, see
[Getting started]({{< relref "./getting-started/_index.md" >}}) or
[Configuration Reference]({{< relref "./configuration/_index.md" >}}) for a detailed reference
on the configuration file.
