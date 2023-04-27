---
title: Debugging
weight: 300
---

# Debugging

Follow these steps to debug issues with Grafana Agent Flow:

1. Use the Grafana Agent Flow UI to debug issues.
2. If the UI doesn't help with debugging an issue, logs can be examined
   instead.

## Grafana Agent Flow UI

Grafana Agent Flow includes an embedded UI viewable from Grafana Agent's HTTP
server, which defaults to listening at `http://localhost:12345`.

> The documentation for the [`grafana-agent run`][grafana-agent run] command describes how to
> modify the address Grafana Agent listens on for HTTP traffic.

### Home page

![](../../../assets/ui_home_page.png)

The home page shows a table of components defined in the config file along with
their health.

Click **View** on a row in the table to navigate to the [Component detail page](#component-detail-page)
for that component.

Click the Grafana Agent logo to navigate back to the home page.

### Graph page

![](../../../assets/ui_graph_page.png)

The **Graph** page shows a graph view of components defined in the config file
along with their health. Clicking a component in the graph navigates to the
[Component detail page](#component-detail-page) for that component.

### Component detail page

![](../../../assets/ui_component_detail_page.png)

The component detail page shows the following information for each component:

* The health of the component with a message explaining the health.
* The current evaluated arguments for the component.
* The current exports for the component.
* The current debug info for the component (if the component has debug info).

> Values marked as a [secret][] are obfuscated and will display as the text
> `(secret)`.

## Debugging using the UI

To debug using the UI:

* Ensure that no component is reported as unhealthy.
* Ensure that the arguments and exports for misbehaving components appear
  correct.

[grafana-agent run]: {{< relref "../reference/cli/run.md" >}}
[secret]: {{< relref "../config-language/expressions/types_and_values.md#secrets" >}}

## Examining logs

Logs may also help debug issues with Grafana Agent Flow.

To reduce logging noise, many components hide debugging info behind debug-level
log lines. It is recommended that you configure the [`logging block`][logging]
to show debug-level log lines when debugging issues with Grafana Agent Flow.

The location of Grafana Agent's logs is different based on how it is deployed.
Refer to the [`logging block`][logging] page to see how to find logs for your
system.

[logging]: {{< relref "../reference/config-blocks/logging.md" >}}

> **NOTE**: NOTE: Please note that the default port for Grafana Agent may vary depending on how it
> is installed. While the default port is traditionally 12345, system packages may override this
> and set the default to a different port. Additionally, certain installations may have different
> default ports. Please refer to your specific installation documentation for the correct default
> port.
