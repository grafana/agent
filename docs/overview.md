# Overview

The Grafana Cloud Agent is an observability data collector optimized for sending
metrics and log data to [Grafana Cloud](https://grafana.com/products/cloud).

The Agent supports collecting Prometheus metrics and Loki logs, both utilizing
the same battle-tested code from the official platforms.

## Metrics

Unlike Prometheus, the Grafana Cloud Agent is _just_ targeting `remote_write`,
so some Prometheus features, such as querying, local storage, recording rules,
and alerts aren't present. `remote_write`, service discovery, and relabeling
rules are included.

The Grafana Cloud Agent has a concept of an "instance", each of which acts as
its own mini Prometheus agent with their own `scrape_configs` section and
`remote_write` rules. More than one instance is useful when you want to have
completely separated configs that write to two different locations without
needing to worry about advanced metric relabeling rules. Multiple instances also
come into play for the [Scraping Service Mode](./scraping-service.md).

The Grafana Cloud Agent can be deployed in three modes:

- Prometheus `remote_write` drop-in
- [Host Filtering mode](#host-filtering)
- [Scraping Service Mode](./scraping-service.md)

The default deployment mode of the Grafana Cloud Agent is the _drop-in_
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

The final mode, _Scraping Service Mode_ is a third operational mode that
clusters a subset of agents. It acts as the in-between of the drop-in mode
(which does no automatic sharding) and `host_filter` mode (which forces sharding
by node). The Scraping Service Mode clusters a set of agents with a set of
shared configs and distributes the scrape load automatically between them. For
more information, please read the dedicated
[Scraping Service Mode](./scraping-service.md) documentation.

### Host Filtering

Host Filtering currently works best with Kubernetes Service Discovery. It does
the following:

1. Gets the hostname of the agent by the `HOSTNAME` environment variable or
   through the default.
2. Checks if the hostname of the agent matches the label value for `__address__`
   or `__meta_kubernetes_pod_node_name` on the discovered target.

If the filter passes, the target is allowed to be scraped. Otherwise, the target
will be silently ignored and not scraped.

## Logs

Grafana Cloud Agent supports collecting logs and sending them to Loki using its
`loki` subsystem. This is done by utilizing the upstream
[Promtail](https://grafana.com/docs/loki/latest/clients/promtail/) client, which
is the official first-party log collection client created by the Loki
developer team.

## Traces

Grafana Cloud Agent supports collecting traces and sending them to Tempo using its
`tempo` subsystem. This is done by utilizing the upstream [OpenTelmetry Collector](https://github.com/open-telemetry/opentelemetry-collector).
The agent is capable of ingesting OpenTelemetry, OpenCensus, Jaeger or Zipkin spans.
See documentation on how to configure [receivers](./configuration-reference.md#tempo_config).
The agent is capable of exporting to any OpenTelemetry GRPC compatible system.

## Comparison to Alternatives

Grafana Cloud Agent is custom built for [Grafana Cloud](https://grafana.com/products/cloud/),
but can be used while using an on-prem `remote_write`-compatible Prometheus API
and an on-prem Loki. Unlike alternatives, Grafana Cloud Agent extends the
official code with extra functionality. This allows the Agent to give an
experience closest to its official counterparts compared to alternatives which
may try to reimplement everything from scratch.

## Next Steps

For more information on installing and running the agent, see
[Getting started](./getting-started.md) or
[Configuration Reference](./configuration-reference.md) for a detailed reference
on the configuration file.

