---
aliases:
- /docs/grafana-cloud/agent/flow/setup/deploy-agent/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/deploy-agent/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/deploy-agent/
- /docs/grafana-cloud/send-data/agent/flow/setup/deploy-agent/
canonical: https://grafana.com/docs/agent/latest/flow/setup/start-agent/
description: Learn about possible deployment topologies for Grafana Agent Flow
menuTitle: Deploy Grafana Agent Flow
title: Grafana Agent Flow deployment topologies
weight: 900
---

{{< docs/shared source="agent" lookup="/deploy-agent.md" version="<AGENT_VERSION>" >}}

## Scaling Grafana Agent

If the load on the Agents is small, it is recommended to process
all necessary telemetry signals in the same Agent process. For example, 
a single Agent can process all of the incoming metrics, logs, traces, and profiles.

However, if the load on the Agents is big, it may be beneficial to
process different telemetry signals in different deployments of Agents:
* This provides better stability due to the isolation between processes.
  * For example, an overloaded Agent processing traces won't impact an Agent processing metrics.
* Different types of signal collection require different methods for scaling:
  * "Pull" components such as `prometheus.scrape` and `pyroscope.scrape` are scaled using hashmod sharing or clustering. 
  * "Push" components such as `otelcol.receiver.otlp` are scaled by placing a load balancer in front of them.

### Traces

<!-- TODO: Link to https://opentelemetry.io/docs/collector/scaling/ ? -->

#### When to scale

<!-- 
TODO: Include information from https://opentelemetry.io/docs/collector/scaling/#when-to-scale
Unfortunately the Agent doesn't have many of the metrics they mention because they're instrumented with OpenCensus and not OpenTelemetry.
-->

#### Stateful and stateless components

In the context of tracing, a "stateful component" is a component 
which needs to aggregate certain spans in order to work correctly.
A "stateless Agent" is an Agent which does not contain stateful components.

Scaling stateful Agents is more difficult, because spans must be forwarded to a 
specific Agent according to a span property such as trace ID or a `service.name` attribute.
This can be done using `otelcol.exporter.loadbalancing`.

Examples of stateful components:

* `otelcol.processor.tail_sampling`
* `otelcol.connector.spanmetrics`
* `otelcol.connector.servicegraph`

<!-- TODO: link to the otelcol.exporter.loadbalancing docs for more info -->

A "stateless component" does not need to aggregate specific spans in 
order to work correctly - it can work correctly even if it only has 
some of the spans of a trace.

Stateless Agents can be scaled without using `otelcol.exporter.loadbalancing`.
You could use an off-the-shelf load balancer to, for example, do a round-robin load balancing.

Examples of stateless components:
* `otelcol.processor.probabilistic_sampler`
* `otelcol.processor.transform`
* `otelcol.processor.attributes`
* `otelcol.processor.span`
