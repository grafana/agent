---
title: Start
weight: 700
---

# Start Grafana Agent Flow

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
