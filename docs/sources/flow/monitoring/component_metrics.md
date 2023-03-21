---
aliases:
- component-metrics/
title: Component metrics
weight: 200
---

# Component metrics

Grafana Agent Flow [components][] may optionally expose Prometheus metrics
which can be used to investigate the behavior of that component. These
component-specific metrics are only generated when an instance of that
component is running.

> Component-specific metrics are different than any metrics being processed by
> the component. Component-specific metrics are used to expose the state of a
> component for observability, alerting, and debugging.

Component-specific metrics are exposed at the `/metrics` HTTP endpoint of the
Grafana Agent HTTP server, which defaults to listening on
`http://localhost:12345`.

> The documentation for the [`grafana-agent run`][grafana-agent run] command describes how to
> modify the address Grafana Agent listens on for HTTP traffic.

Component-specific metrics will have a `component_id` label matching the
component ID generating those metrics. For example, component-specific metrics
for a `prometheus.remote_write` component labeled `production` will have a
`component_id` label with the value `prometheus.remote_write.production`.

The [reference documentation][] for each component will describe the list of
component-specific metrics that component exposes. Not all components will
expose metrics.

[components]: {{< relref "../concepts/components.md" >}}
[grafna-agent run]: {{< relref "../reference/cli/run.md" >}}
[reference documentation]: {{< relref "../reference/components/_index.md" >}}
