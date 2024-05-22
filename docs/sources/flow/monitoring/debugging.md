---
aliases:
- /docs/grafana-cloud/agent/flow/monitoring/debugging/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/monitoring/debugging/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/monitoring/debugging/
- /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/
canonical: https://grafana.com/docs/agent/latest/flow/monitoring/debugging/
description: Learn about debugging
title: Debugging
weight: 300
refs:
  clustering:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/concepts/clustering/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/concepts/clustering/
  logging:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/logging/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/logging/
  grafana-agent-run:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/cli/run/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/cli/run/
  secret:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/config-language/expressions/types_and_values/#secrets.md
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/config-language/expressions/types_and_values/#secrets.md
  install:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/setup/install/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/setup/install/
---

# Debugging

Follow these steps to debug issues with {{< param "PRODUCT_NAME" >}}:

1. Use the {{< param "PRODUCT_NAME" >}} UI to debug issues.
2. If the {{< param "PRODUCT_NAME" >}} UI doesn't help with debugging an issue, logs can be examined
   instead.

## {{% param "PRODUCT_NAME" %}} UI

{{< param "PRODUCT_NAME" >}} includes an embedded UI viewable from the {{< param "PRODUCT_ROOT_NAME" >}} HTTP
server, which defaults to listening at `http://localhost:12345`.

> **NOTE**: For security reasons, installations of {{< param "PRODUCT_NAME" >}} on
> non-containerized platforms default to listening on `localhost`. default
> prevents other machines on the network from being able to view the UI.
>
> To expose the UI to other machines on the network on non-containerized
> platforms, refer to the documentation for how you [installed](ref:install)
> {{< param "PRODUCT_NAME" >}}.
>
> If you are running a custom installation of {{< param "PRODUCT_NAME" >}}, refer to the
> documentation for [the `grafana-agent run` command][grafana-agent run] to
> learn how to change the HTTP listen address, and pass the appropriate flag
> when running {{< param "PRODUCT_NAME" >}}.

### Home page

![](../../../assets/ui_home_page.png)

The home page shows a table of components defined in the configuration file along with
their health.

Click **View** on a row in the table to navigate to the [Component detail page](#component-detail-page)
for that component.

Click the {{< param "PRODUCT_ROOT_NAME" >}} logo to navigate back to the home page.

### Graph page

![](../../../assets/ui_graph_page.png)

The **Graph** page shows a graph view of components defined in the configuration file
along with their health. Clicking a component in the graph navigates to the
[Component detail page](#component-detail-page) for that component.

### Component detail page

![](../../../assets/ui_component_detail_page.png)

The component detail page shows the following information for each component:

* The health of the component with a message explaining the health.
* The current evaluated arguments for the component.
* The current exports for the component.
* The current debug info for the component (if the component has debug info).

> Values marked as a [secret](ref:secret) are obfuscated and will display as the text
> `(secret)`.

### Clustering page

![](../../../assets/ui_clustering_page.png)

The clustering page shows the following information for each cluster node:

* The node's name.
* The node's advertised address.
* The node's current state (Viewer/Participant/Terminating).
* The local node that serves the UI.

## Debugging using the UI

To debug using the UI:

* Ensure that no component is reported as unhealthy.
* Ensure that the arguments and exports for misbehaving components appear
  correct.

## Examining logs

Logs may also help debug issues with {{< param "PRODUCT_NAME" >}}.

To reduce logging noise, many components hide debugging info behind debug-level
log lines. It is recommended that you configure the [`logging` block](ref:logging)
to show debug-level log lines when debugging issues with {{< param "PRODUCT_NAME" >}}.

The location of {{< param "PRODUCT_NAME" >}} logs is different based on how it is deployed.
Refer to the [`logging` block](ref:logging) page to see how to find logs for your
system.

## Debugging clustering issues

To debug issues when using [clustering](ref:clustering), check for the following symptoms.

- **Cluster not converging**: The cluster peers are not converging on the same
  view of their peers' status. This is most likely due to network connectivity
issues between the cluster nodes. Use the Flow UI of each running peer to
understand which nodes are not being picked up correctly.
- **Cluster split brain**: The cluster peers are not aware of one another,
  thinking they’re the only node present. Again, check for network connectivity
issues. Check that the addresses or DNS names given to the node to join are
correctly formatted and reachable.
- **Configuration drift**: Clustering assumes that all nodes are running with
  the same configuration file at roughly the same time. Check the logs for
issues with the reloaded configuration file as well as the graph page to verify
changes have been applied.
- **Node name conflicts**: Clustering assumes all nodes have unique names;
  nodes with conflicting names are rejected and will not join the cluster. Look
at the clustering UI page for the list of current peers with their names, and
check the logs for any reported name conflict events.
- **Node stuck in terminating state**: The node attempted to gracefully shut
down and set its state to Terminating, but it has not completely gone away. Check
the clustering page to view the state of the peers and verify that the
terminating {{< param "PRODUCT_ROOT_NAME" >}} has been shut down.


