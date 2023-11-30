---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/start-agent/
- /docs/grafana-cloud/send-data/agent/static/set-up/start-agent/
canonical: https://grafana.com/docs/agent/latest/static/set-up/start-agent/
description: Learn how to start, restart, and stop Grafana Agent in static mode
menuTitle: Start static mode
title: Start, restart, and stop Grafana Agent in static mode
weight: 200
---

# Start, restart, and stop Grafana Agent in static mode

You can start, restart, and stop Grafana Agent after it is installed.

## Linux

Grafana Agent is installed as a [systemd][] service on Linux.

[systemd]: https://systemd.io/

### Start Grafana Agent

To start Grafana Agent, run the following command in a terminal window:

```shell
sudo systemctl start grafana-agent
```

(Optional) Verify that the service is running:

```shell
sudo systemctl status grafana-agent
```

### Configure Grafana Agent to start at boot

To automatically run Grafana Agent when the system starts, run the following command in a terminal window:

```shell
sudo systemctl enable grafana-agent.service
```

### Restart Grafana Agent

To restart Grafana Agent, run the following command in a terminal window:

```shell
sudo systemctl restart grafana-agent
```

### Stop Grafana Agent

To stop Grafana Agent, run the following command in a terminal window:

```shell
sudo systemctl stop grafana-agent
```

### View Grafana Agent logs on Linux

To view the Grafana Agent log files, run the following command in a terminal window:

```shell
sudo journalctl -u grafana-agent
```

## macOS

Grafana Agent is installed as a launchd service on macOS.

### Start Grafana Agent

To start Grafana Agent, run the following command in a terminal window:

```shell
brew services start grafana-agent
```

Grafana Agent automatically runs when the system starts.

Optional: Verify that the service is running:

```shell
brew services info grafana-agent
```

### Restart Grafana Agent

To restart Grafana Agent, run the following command in a terminal window:

```shell
brew services restart grafana-agent
```

### Stop Grafana Agent

To stop Grafana Agent, run the following command in a terminal window:

```shell
brew services stop grafana-agent
```

### View Grafana Agent logs on macOS

By default, logs are written to `$(brew --prefix)/var/log/grafana-agent.log` and
`$(brew --prefix)/var/log/grafana-agent.err.log`.

If you followed [Configure][configure] steps in the macOS install instructions and changed the path where logs are written, refer to your current copy of the Grafana Agent formula to locate your log files.

## Windows

Grafana Agent is installed as a Windows Service. The service is configured to automatically run on startup.

To verify that Grafana Agent is running as a Windows Service:

1. Open the Windows Services manager (services.msc):

   1. Right click on the Start Menu and select **Run**.

   1. Type: `services.msc` and click **OK**.

1. Scroll down to find the **Grafana Agent** service and verify that the **Status** is **Running**.

### View Grafana Agent logs

When running on Windows, Grafana Agent writes its logs to Windows Event
Logs with an event source name of **Grafana Agent**.

To view the logs, perform the following steps:

1. Open the Event Viewer:

   1. Right click on the Start Menu and select **Run**.

   1. Type `eventvwr` and click **OK**.

1. In the Event Viewer, click on **Windows Logs > Application**.

1. Search for events with the source **Grafana Agent**.

## Standalone binary

If you downloaded the standalone binary, you must run the agent from a terminal or command window.

### Start Grafana Agent on Linux, macOS, FreeBSD, or Windows

To start Grafana Agent on Linux, macOS, FreeBSD, or Windows run the following command in a terminal window or command prompt:

```shell
BINARY_PATH -config.file CONFIG_FILE 
```

Replace the following:

* `BINARY_PATH`: The path to the Grafana Agent binary file
* `CONFIG_FILE`: The path to the Grafana Agent configuration file.

{{% docs/reference %}}
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-macos#configure"
[configure]: "/docs/grafana-cloud/ -> ./install/install-agent-macos/#configure"
{{% /docs/reference %}}
