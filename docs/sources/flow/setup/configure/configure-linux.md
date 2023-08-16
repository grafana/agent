---
canonical: https://grafana.com/docs/agent/latest/flow/setup/configure/configure-linux/
description: Learn how to configure Grafana Agent in flow mode on Linux
menuTitle: Linux
title: Configure Grafana Agent in flow mode on Linux
weight: 300
---

# Configure Grafana Agent in flow mode on Linux

To configure Grafana Agent in flow mode on Linux, perform the following steps:

1. Edit the default configuration file at `/etc/grafana-agent-flow.river`.

1. Run the following command in a terminal to reload the configuration file:

   ```shell
   sudo systemctl reload grafana-agent-flow
   ```

To change the configuration file used by the service, perform the following steps:

1. Edit the environment file for the service:

   * Debian or Ubuntu: edit `/etc/default/grafana-agent-flow`
   * RHEL/Fedora or SUSE/openSUSE: edit `/etc/sysconfig/grafana-agent-flow`

1. Change the contents of the `CONFIG_FILE` environment variable to point to
   the new configuration file to use.

1. Restart the Grafana Agent service:

   ```shell
   sudo systemctl restart grafana-agent-flow
   ```

## Pass additional command-line flags

By default, the Grafana Agent service launches with the [run][]
command, passing the following flags:

* `--storage.path=/var/lib/grafana-agent-flow`

To pass additional command-line flags to the Grafana Agent binary, perform
the following steps:

1. Edit the environment file for the service:

   * Debian-based systems: edit `/etc/default/grafana-agent-flow`
   * RedHat or SUSE-based systems: edit `/etc/sysconfig/grafana-agent-flow`

1. Change the contents of the `CUSTOM_ARGS` environment variable to specify
   command-line flags to pass.

1. Restart the Grafana Agent service:

   ```shell
   sudo systemctl restart grafana-agent-flow
   ```

To see the list of valid command-line flags that can be passed to the service,
refer to the documentation for the [run][] command.

[run]: {{< relref "../../reference/cli/run.md" >}}

## Expose the UI to other machines

By default, Grafana Agent listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Pass additional command-line flags](#pass-additional-command-line-flags)
   to edit command line flags passed to Grafana Agent, including the
   following customizations:

    1. Add the following command line argument to `CUSTOM_ARGS`:

       ```shell
       --server.http.listen-addr=LISTEN_ADDR:12345
       ```

       Replace `LISTEN_ADDR` with an address which other machines on the
       network have access to, like the network IP address of the machine
       Grafana Agent is running on.

       To listen on all interfaces, replace `LISTEN_ADDR` with `0.0.0.0`.

[UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
