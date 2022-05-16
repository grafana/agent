---
aliases:
- /docs/agent/latest/
title: Grafana Agent
weight: 1
---

# Grafana Agent

Grafana Agent collects and forwards telemetry data to open source deployments of the LGTM (Loki, Grafana, Tempo, Mimir) Stack, Grafana Cloud, or Grafana Enterprise, where your data can then be analyzed. You can install Grafana Agent on Kubernetes and Docker, or as a system process for Linux, macOS, and Windows machines.  

The Grafana Agent is open source and its source code is available on GitHub at https://github.com/grafana/agent.

Grafana Agent is for engineers, operators, or administrators who want to collect and forward telemetry data for analysis and on-call alerting. Those operating Grafana Agent must install and configure Grafana Agent to properly collect telemetry data and monitor the health of running agents.

## Features

There are other ways of sending metrics, logs and traces to the LGTM Stack, Grafana Cloud or Grafana Enterprise, but there are a few advantages of using the Grafana Agent. These features are outlined below.

* Provides a one-stop solution for collecting metrics, logs, and traces.
* Collects out-of-the-box telemetry from popular projects like MySQL through integrations
* Works seamlessly with the LGTM Stack. Alternatively, metrics can be sent to any Prometheus-compatible endpoint, and traces can be sent to any OTLP-compatible endpoint.  
* Offers new solutions to help scale metrics collection like host_filtering and sharding 
* Provides the Grafana Agent Operator, which enables individual teams to manage their configurations through PodMonitors, ServiceMonitors, and Probes.

## Metrics

Grafana Agent focuses metrics support around Prometheus' remote_write protocol,
so some Prometheus features, such as querying, local storage, recording rules,
and alerts are not present. `remote_write`, service discovery, and relabeling
rules are included.

The Grafana Agent has a concept of an "instance" each of which acts as
its own mini Prometheus agent with its own `scrape_configs` section and
`remote_write` rules. More than one instance is useful when you want to have
separate configurations that write to two different locations without
needing to consider advanced metric relabeling rules. Multiple instances also
come into play for the [Scraping Service Mode]({{< relref "./scraping-service" >}}).

The Grafana Agent for collecting metrics can be deployed in three modes:

- Prometheus `remote_write` drop-in
- [Host Filtering mode](#host-filtering)
- [Scraping Service mode]({{< relref "./scraping-service" >}})

### Prometheus `remote_write` drop-in
The default deployment mode of the Grafana Agent is a _drop-in_
replacement for Prometheus `remote_write`. The Agent acts similarly to a
single-process Prometheus, doing service discovery, scraping, and remote
writing.


### Host filtering
Host filtering configures Agents to scrape targets that are running on the same
machine as the Grafana Agent process.

1. Gets the hostname of the agent by the `HOSTNAME` environment variable or
   through the default.
2. Checks if the hostname of the agent matches the label value for `__address__`
   service-discovery-specific node labels against the discovered target.

If the filter passes, the target is scraped. Otherwise, the target
is ignored and not scraped.

To use _Host Filtering mode_, you set a `host_filter` flag on a specific
instance inside the Agent's configuration file. When you set this flag, the
instance only scrapes metrics from targets that are running on the same
machine. This is useful for migrating to sharded
Prometheus instances in a Kubernetes cluster, where the Agent can be deployed as
a DaemonSet and distribute memory requirements across multiple nodes.

Note that _Host Filtering_ mode and sharding your instances means that if an
Agent's metrics are being sent to an alerting system, alerts for that Agent might
not be able to be generated if the entire node has problems. This changes the
semantics of failure detection, and alerts would have to be configured to catch
agents not reporting in.


For more information on the host filtering mode, refer to the [operation
guide]({{< relref "./operation-guide#host-filtering" >}}).

### Scraping Service 
_Scraping Service Mode_ 
clusters a subset of agents. It acts as a go-between for the drop-in mode
(which does no automatic sharding) and `host_filter` mode (which forces sharding
by node). The Scraping Service Mode clusters a set of agents with a set of
shared configurations and distributes the scrape load automatically between them. For
more information on Scraping Service, see [Scraping Service]({{< relref "./scraping-service" >}}).

## Logs

Grafana Agent supports collecting logs and sending them to Loki using its
`loki` subsystem. This is done using the upstream
[Promtail](https://grafana.com/docs/loki/latest/clients/promtail/) client, which
is the official first-party log collection client created by the Loki
developer team.

## Traces

Grafana Agent collects traces and forwards them to Tempo using its
`traces` subsystem. This is done using the upstream [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).
Agent can ingest OpenTelemetry, OpenCensus, Jaeger, Zipkin, or Kafka spans.
For more information on how to configure, refer to [receivers]({{< relref "./configuration/traces-config.md" >}}).
The agent is capable of exporting to any OpenTelemetry GRPC compatible system.

