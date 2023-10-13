---
aliases:
- /docs/grafana-cloud/agent/flow/monitoring/controller_metrics/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/monitoring/controller_metrics/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/monitoring/controller_metrics/
- controller-metrics/
canonical: https://grafana.com/docs/agent/latest/flow/monitoring/controller_metrics/
title: Controller metrics
description: Learn about controller metrics
weight: 100
---

# Controller metrics

The Grafana Agent Flow [component controller][] exposes Prometheus metrics
which can be used to investigate the controller state.

Metrics for the controller are exposed at the `/metrics` HTTP endpoint of the
Grafana Agent HTTP server, which defaults to listening on
`http://localhost:12345`.

> The documentation for the [`grafana-agent run`][grafana-agent run] command
> describes how to modify the address Grafana Agent listens on for HTTP
> traffic.

The controller exposes the following metrics:

* `agent_component_controller_evaluating` (Gauge): Set to `1` whenever the
  component controller is currently evaluating components. Note that this value
  may be misrepresented depending on how fast evaluations complete or how often
  evaluations occur.
* `agent_component_controller_running_components` (Gauge): The current
  number of running components by health. The health is represented in the
  `health_type` label.
* `agent_component_evaluation_seconds` (Histogram): The time it takes to 
  evaluate components after one of their dependencies is updated.
* `agent_component_dependencies_wait_seconds` (Histogram): Time spent by 
  components waiting to be evaluated after one of their dependencies is updated.
* `agent_component_evaluation_queue_size` (Gauge): The current number of
  component evaluations waiting to be performed.

[component controller]: {{< relref "../concepts/component_controller.md" >}}
[grafana-agent run]: {{< relref "../reference/cli/run.md" >}}
