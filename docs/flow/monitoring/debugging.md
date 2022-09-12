---
aliases:
- /docs/agent/latest/flow/monitoring/debugging
title: Debugging
weight: 300
---

# Debugging

There are two primary ways to debug issues with Grafana Agent Flow:

1. Through the Grafana Agent Flow UI
2. Through examining logs

## Grafana Agent Flow UI

Grafana Agent Flow includes an embedded UI viewable from Grafana Agent's HTTP
server, which defaults to listening at `http://localhost:12345`.

> The documentation for the [`agent run`][agent run] command describes how to
> modify the address Grafana Agent listens on for HTTP traffic.

This UI provides two views:

* The home page shows a table of running components along with their health.
  Clicking the "View" button navigates to the detail page for that component.
* The "Graph" page shows a graph view of all running components along with
  their health. Clicking a component in the graph navigates to the detail page
  for that component.

The component detail page shows the following information for each component:

* The health of the component with a message explaining the health.
* The current evaluated arguments for the component.
* The current exports for the component.
* The current debug info for the component (if the component has debug info).

To debug using the UI:

* Ensure that no component is reported as unhealthy.
* Ensure that the arguments and exports for misbehaving components appear
  correct.

[agent run]: {{< relref "../reference/cli/run.md" }}

## Examining logs

Logs may also help debug issues with Grafana Agent Flow.

To reduce logging noise, many components hide debugging info behind debug-level
log lines. It is recommended that you configure the [`logging block`][logging]
to show debug-level log lines when debugging issues with Grafana Agent Flow.

The location of Grafana Agent's logs is different based on how it is deployed.
Refer to the [`logging block`][logging] page to see how to find logs for your
system.

[logging]: {{< relref "../reference/config-blocks/logging.md" >}}
