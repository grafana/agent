# Overview

The Grafana Cloud Agent is an observability data collector optimized for sending
metrics and log data to [Grafan Cloud](https://grafana.com/products/cloud).

Currently, it only comes with support for collecting and sending Prometheus
metrics, accomplished through utilizing the same battle-tested code that
Prometheus contains.

Unlike Prometheus, the Grafana Cloud Agent is _just_ targeting `remote_write`,
so some Prometheus features, such as querying, local storage, recording rules,
and alerts aren't present. `remote_write`, service discovery, and relabeling
rules are included.

The Grafana Cloud Agent has a concept of an "instance", each of which acts as
its own mini Prometheus agent with own `scrape_configs` section and
`remote_write` rules. Most users will only ever need to define one instance.
Multiple instances will be more useful in the future when a clustering mode is
added to the Agent.

The Grafana Cloud Agent can be deployed in two modes:

- Prometheus `remote_write` drop-in
- Host Filtering mode

The default deployment mode of the Grafana Cloud Agent is the drop-in
replacement for Prometheus `remote_write`. The Agent will act similarly to a
single-processed Prometheus, doing service discovery, scraping, and remote
writing.

The other deployment mode, Host Filtering mode, is achieved by setting a
`host_filter` flag on a specific instance inside the Agent's configuration file.
When this flag is set, the instance will only scrape metrics from targets that
are running on the same machine as the target. This is extremely useful to
migrate to sharded Prometheus instances in a Kubernetes cluster, where the Agent
can then be deployed as a DaemonSet and distribute memory requirements across
multiple nodes.

Note that Host Filtering mode and sharding your instances means that if an
Agent's metrics are being sent to an alerting system, alerts for that Agent may
not be able to be generated if the entire node has problems. This changes the
semantics of failure detection, and alerts would have to be configured to catch
agents not reporting in.

For more information on installing and running the agent, see
[Getting started](./getting-started.md) or
[Configuration Reference](./configuration-reference.md) for a detailed reference
on the configuration file.

## Host Filtering

Host Filtering currently works best with Kubernetes Service Discovery. It does
the following:

1. Gets the hostname of the agent by the `HOSTNAME` environment variable or
   through the default.
2. Checks if the hostname of the agent matches the label value for `__address__`
   or `__meta_kubernetes_pod_node_name` on the discovered target.

If the filter passes, the target is allowed to be scraped. Otherwise, the target
will be silently ignored and not scraped.

## Comparison to Alternatives

Grafana Cloud Agent aims to give an experience closest to Prometheus, by
providing Prometheus features like service discovery, meta labels, and
relabeling. This is primarily achieved by the Agent vendoring Prometheus an
using its code.

Alternatives that support Prometheus metrics try to incorporate more than just
Prometheus metrics ingestion, and tend to reimplement the code for doing so.
This leads to missing features or the other agents feeling like a
jack-of-all-trades, master-of-none.

## Roadmap

Today, the Grafana Cloud Agent can only scrape Prometheus metrics. In the
future, we are planning on adding support for data corresponding to the other
Grafana Cloud hosted platforms:

- Graphite metrics
- Loki logs

Operationally, we are also planning on adding a distributed scraping service
mode, where the Agent could be deployed as a cluster. This will be the third
deployment mechanism supported, outside of the currently supported
single-process and DaemonSet modes.
