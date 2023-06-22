---
title: Configure on macOS
weight: 300
---

# Configure Grafana Agent Flow on macOS

To configure Grafana Agent Flow on macOS, perform the following
steps:

1. Edit the default configuration file at
   `$(brew --prefix)/etc/grafana-agent-flow/config.river`.

2. Run the following command in a terminal to restart the Grafana Agent Flow
   service:

   ```shell
   brew services restart grafana-agent-flow
   ```

## Configure the Grafana Agent Flow service

> **NOTE**: Due to limitations in Homebrew, customizing the service used by
> Grafana Agent Flow on macOS requires changing the Homebrew formula and
> reinstalling Grafana Agent Flow.

To customize the Grafana Agent Flow service on macOS, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   brew edit grafana-agent-flow
   ```

   This will open the Grafana Agent Flow Homebrew Formula in an editor.

2. Modify the `service` section as desired to change things such as:

   * The configuration file used by Grafana Agent Flow.
   * Flags passed to the Grafana Agent Flow binary.
   * Location of log files.

   When done, save the resulting formula file.

3. Reinstall the Grafana Agent Flow Formula by running the following command in
   a terminal:

   ```shell
   brew reinstall grafana-agent-flow
   ```

4. Restart the Grafana Agent Flow service by running the command in a terminal:

   ```shell
   brew services restart grafana-agent-flow
   ```

## Expose the UI to other machines

By default, Grafana Agent Flow listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Configuring the Grafana Agent Flow service](#configuring-the-grafana-agent-flow-service)
   to edit command line flags passed to Grafana Agent Flow, including the
   following customizations:

    1. Modify the line inside the `service` block containing
       `--server.http.listen-addr=127.0.0.1:12345`, replacing `127.0.0.1` with
       the address which other machines on the network have access to, like the
       network IP address of the machine Grafana Agent Flow is running on.

       To listen on all interfaces, replace `127.0.0.1` with `0.0.0.0`.

[UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
