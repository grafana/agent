---
title: macOS
weight: 400
---

# Install Grafana Agent Flow on macOS

Grafana Agent Flow can be installed on Intel and Apple Silicon macOS machines.

## Before you begin

* Ensure that [Homebrew][] is installed on your machine.

[Homebrew]: https://brew.sh

## Install

To install Grafana Agent Flow on macOS, perform the following steps:

1. Add the Grafana Homebrew tap by running the following command in a terminal
   window:

   ```shell
   brew tap grafana/grafana
   ```

2. Install the Grafana Agent Flow formula by running the folloing command in a
   terminal window:

   ```shell
   brew install grafana-agent-flow
   ```

## Operation guide

After installing Grafana Agent Flow on macOS, it will be exposed as a
launchd service.

### Run Grafana Agent Flow

To run Grafana Agent Flow, run the following command in a terminal:

```shell
brew services start grafana-agent-flow
```

After running this command, Grafana Agent Flow will be running as a service and
will automatically run on system startup.

To check the status of Grafana Agent Flow, run the following command in a
terminal:

```shell
brew services info grafana-agent-flow
```

### Configuring Grafana Agent Flow

To configure Grafana Agent Flow when installed on macOS, perform the following
steps:

1. Edit the default configuration file at
   `$(brew --prefix)/etc/grafana-agent-flow/config.river`.

2. Run the following command in a terminal to restart the Grafana Agent Flow
   service:

   ```shell
   brew services restart grafana-agent-flow
   ```

### Configuring the Grafana Agent Flow service

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

4. Reatart the Grafana Agent Flow service by running the command in a terminal:

   ```shell
   brew services restart grafana-agent-flow
   ```

### Viewing Grafana Agent Flow logs

By default, logs are written to `$(brew --prefix)/var/log/grafana-agent.log` and
`$(brew --preifx)/var/log/grafana-agent-flow.err.log`.

If you followed [Configuring the Grafana Agent Flow service](#configuring-the-grafana-agent-flow-service)
and changed the path where logs are written, refer to your current copy of the
Grafana Agent Flow formula to locate log files.
