---
aliases:
  - /docs/grafana-cloud/agent/flow/setup/start/linux/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/start/linux/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/start/linux/
  - /docs/grafana-cloud/send-data/agent/flow/setup/start/linux/
canonical: https://grafana.com/docs/agent/latest/flow/setup/start/linux/
description: Learn how to start Grafana Agent Flow on Linux
menuTitle: Linux
title: Start Grafana Agent Flow on Linux
weight: 300
---

# Start {{% param "PRODUCT_NAME" %}} on Linux

{{< param "PRODUCT_NAME" >}} is [installed][InstallLinux] as a [systemd][] service on Linux.

[systemd]: https://systemd.io/

## Start {{% param "PRODUCT_NAME" %}}

To start {{< param "PRODUCT_NAME" >}}, run the following command in a terminal window:

```shell
sudo systemctl start grafana-agent-flow
```

(Optional) To verify that the service is running, run the following command in a terminal window:

```shell
sudo systemctl status grafana-agent-flow
```

## Configure {{% param "PRODUCT_NAME" %}} to start at boot

To automatically run {{< param "PRODUCT_NAME" >}} when the system starts, run the following command in a terminal window:

```shell
sudo systemctl enable grafana-agent-flow.service
```

## Restart {{% param "PRODUCT_NAME" %}}

To restart {{< param "PRODUCT_NAME" >}}, run the following command in a terminal window:

```shell
sudo systemctl restart grafana-agent-flow
```

## Stop {{% param "PRODUCT_NAME" %}}

To stop {{< param "PRODUCT_NAME" >}}, run the following command in a terminal window:

```shell
sudo systemctl stop grafana-agent-flow
```

## View {{% param "PRODUCT_NAME" %}} logs on Linux

To view {{< param "PRODUCT_NAME" >}} log files, run the following command in a terminal window:

```shell
sudo journalctl -u grafana-agent-flow
```

## Next steps

- [Configure {{< param "PRODUCT_NAME" >}}][Configure]

{{% docs/reference %}}
[InstallLinux]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/setup/install/linux.md"
[InstallLinux]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/flow/setup/install/linux.md"
[Configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/configure/configure-linux.md"
[Configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-linux.md"
{{% /docs/reference %}}
