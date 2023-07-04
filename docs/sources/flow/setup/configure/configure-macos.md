---
description: Learn how to configure Grafana Agent in flow mode on macOS
title: Configure Grafana Agent in flow mode on macOS
menuTitle: macOS
weight: 400
---

# Configure Grafana Agent on macOS

To configure Grafana Agent in flow mode on macOS, perform the following steps:

1. Edit the default configuration file at `$(brew --prefix)/etc/grafana-agent-flow/config.river`.

1. Run the following command in a terminal to restart the Grafana Agent service:

   ```shell
   brew services restart grafana-agent-flow
   ```

## Configure the Grafana Agent service

{{% admonition type="note" %}}
Due to limitations in Homebrew, customizing the service used by
Grafana Agent on macOS requires changing the Homebrew formula and
reinstalling Grafana Agent.
{{% /admonition %}}

To customize the Grafana Agent service on macOS, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   brew edit grafana-agent-flow
   ```

   This will open the Grafana Agent Homebrew Formula in an editor.

1. Modify the `service` section as desired to change things such as:

   * The River configuration file used by Grafana Agent.
   * Flags passed to the Grafana Agent binary.
   * Location of log files.

   When you are done, save the file.

1. Reinstall the Grafana Agent Formula by running the following command in a terminal:

   ```shell
   brew reinstall grafana-agent-flow
   ```

1. Restart the Grafana Agent service by running the command in a terminal:

   ```shell
   brew services restart grafana-agent-flow
   ```

## Expose the UI to other machines

By default, Grafana Agent listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Configure the Grafana Agent service](#configure-the-grafana-agent-flow-service)
   to edit command line flags passed to Grafana Agent, including the
   following customizations:

    1. Modify the line inside the `service` block containing
       `--server.http.listen-addr=127.0.0.1:12345`, replacing `127.0.0.1` with
       the address which other machines on the network have access to, like the
       network IP address of the machine Grafana Agent is running on.

       To listen on all interfaces, replace `127.0.0.1` with `0.0.0.0`.

[UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
