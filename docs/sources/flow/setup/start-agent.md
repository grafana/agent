---
description: Learn how to start, restart, and stop Grafana Agent Flow after it is installed
title: Start, restart, and stop Grafana Agent Flow 
menuTitle: Start
weight: 800
---

# Start Grafana Agent Flow

You can start, restart, and stop Grafana Agent Flow after it is installed.

## Linux

Grafana Agent Flow is configured as a [systemd][] service on Linux.

[systemd]: https://systemd.io/

### Start Grafana Agent Flow

To start Grafana Agent Flow, run the following commands in a terminal window.

1. Start Grafana Agent Flow:

   ```shell
   sudo systemctl start grafana-agent-flow
   ```

1. (Optional) Verify that the service is running:

   ```shell
   sudo systemctl status grafana-agent-flow
   ```

### Configure Grafana Agent Flow to start at boot

To automatically run Grafana Agent Flow when the system starts, run the following command in a terminal:

```shell
sudo systemctl enable grafana-agent-flow.service
```

### Restart Grafana Agent Flow

To restart Grafana Agent Flow, run the following command in a terminal window:

```shell
sudo systemctl restart grafana-agent-flow
```

### Stop Grafana Agent Flow

To stop Grafana Agent Flow, run the following commands in a terminal window:

```shell
sudo systemctl stop grafana-agent-flow
```

### View Grafana Agent Flow logs on Linux

To view the Grafana Agent Flow log files, run the following command in a terminal:

```shell
sudo journalctl -u grafana-agent-flow
```

## Windows

After installing Grafana Agent Flow on Windows, it will be exposed as a Windows
Service, where it automatically runs on startup.

### View Grafana Agent Flow logs

When running on Windows, Grafana Agent Flow writes its logs to Windows Event
Logs with an event source name of "Grafana Agent Flow".

To view the logs, perform the following steps:

1. Open the Event Viewer:
   1. Right click on the Start Menu icon.
   1. Click on **Run**.
   1. In the resulting dialog box, type `eventvwr`.
   1. Click **OK**.
1. In the Event Viewer, click on Windows Logs > Application.
1. Search for events with the source called "Grafana Agent Flow."

## macOS

Grafana Agent Flow is configured as a launchd service on macOS. 

### Start Grafana Agent Flow

1. Start Grafana Agent Flow:

   ```shell
   brew services start grafana-agent-flow
   ```

   Grafana Agent Flow automatically runs when the system starts.

1. (Optional) Verify that the serivce is running:

   ```shell
   brew services info grafana-agent-flow
   ```
### View Grafana Agent Flow logs on macOS

By default, logs are written to `$(brew --prefix)/var/log/grafana-agent-flow.log` and
`$(brew --prefix)/var/log/grafana-agent-flow.err.log`.

If you followed [Configuring the Grafana Agent Flow service](#configuring-the-grafana-agent-flow-service)
and changed the path where logs are written, refer to your current copy of the
Grafana Agent Flow formula to locate log files.
