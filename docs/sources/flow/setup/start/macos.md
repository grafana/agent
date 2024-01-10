---
aliases:
  - /docs/grafana-cloud/agent/flow/setup/start/macos/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/start/macos/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/start/macos/
  - /docs/grafana-cloud/send-data/agent/flow/setup/start/macos/
canonical: https://grafana.com/docs/agent/latest/flow/setup/start/macos/
description: Learn how to start Grafana Agent Flow on macOS
menuTitle: macOS
title: Start Grafana Agent Flow on macOS
weight: 400
---

# Start {{% param "PRODUCT_NAME" %}} on macOS

{{< param "PRODUCT_NAME" >}} is installed as a launchd service on macOS.

## Start {{% param "PRODUCT_NAME" %}}

To start {{< param "PRODUCT_NAME" >}}, run the following command in a terminal window:

```shell
brew services start grafana-agent-flow
```

{{< param "PRODUCT_NAME" >}} automatically runs when the system starts.

(Optional) To verify that the service is running, run the following command in a terminal window:

```shell
brew services info grafana-agent-flow
```

## Restart {{% param "PRODUCT_NAME" %}}

To restart {{< param "PRODUCT_NAME" >}}, run the following command in a terminal window:

```shell
brew services restart grafana-agent-flow
```

## Stop {{% param "PRODUCT_NAME" %}}

To stop {{< param "PRODUCT_NAME" >}}, run the following command in a terminal window:

```shell
brew services stop grafana-agent-flow
```

## View {{% param "PRODUCT_NAME" %}} logs on macOS

By default, logs are written to `$(brew --prefix)/var/log/grafana-agent-flow.log` and
`$(brew --prefix)/var/log/grafana-agent-flow.err.log`.

If you followed [Configure the {{< param "PRODUCT_NAME" >}} service][Configure] and changed the path where logs are written,
refer to your current copy of the {{< param "PRODUCT_NAME" >}} formula to locate your log files.

{{% docs/reference %}}
[Configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/configure/configure-macos.md#configure-the-grafana-agent-service"
[Configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-macos.md#configure-the-grafana-agent-service"
{{% /docs/reference %}}
