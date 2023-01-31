---
aliases:
- controller-metrics/
title: Controller metrics
weight: 100
---

# Controller metrics

The Grafana Agent Flow [component controller][] exposes Prometheus metrics
which can be used to investigate the controller state.

Metrics for the controller are exposed at the `/metrics` HTTP endpoint of the
Grafana Agent HTTP server, which defaults to listening on
`http://localhost:12345`.

> The documentation for the [`agent run`][agent run] command describes how to
> modify the address Grafana Agent listens on for HTTP traffic.

The controller exposes the following metrics:

* `agent_component_controller_evaluating` (Gauge): Set to `1` whenever the
  component controller is currently evaluating components. Note that this value
  may be misrepresented depending on how fast evaluations complete or how often
  evaluations occur.
* `agent_component_controller_running_components_total` (Gauge): The current
  number of running components by health. The health is represented in the
  `health_type` label.
* `agent_component_evaluation_seconds` (Histogram): The number of completed
  graph evaluations performed by the component controller with how long they
  took.

[component controller]: {{< relref "../concepts/component_controller.md" >}}
[agent run]: {{< relref "../reference/cli/run.md" >}}
