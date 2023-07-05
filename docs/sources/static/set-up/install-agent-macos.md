---
title: Install Grafana Agent in static mode on macOS
menuTitle: macOS
weight: 500
aliases:
- ../../set-up/install-agent-macos/
---

# Install static mode on macOS

You can install Grafana Agent in static mode on macOS.

## Before you begin

Ensure that [Homebrew][] is installed.

{{% admonition type="note" %}}
The default prefix for Homebrew on Intel is `/usr/local`. The default prefix for Homebrew on Apple Silicon is `/opt/Homebrew`. You can verify the default prefix for Homebrew on your computer by opening a terminal and typing `brew --prefix`.
{{% /admonition %}}

[Homebrew]: https://brew.sh

## Installing Grafana Agent with Homebrew

Open a terminal and run the following commands:

```shell
brew update
brew install grafana-agent
```

   Grafana Agent is installed by default at `$(brew --prefix)/Cellar/grafana-agent/VERSION`.

## Configuring Grafana Agent

1. To create the Agent `config.yml` file, open a terminal and run the following command:

    ```shell
    touch $(brew --prefix)/etc/grafana-agent/config.yml
    ```

1. Edit `$(brew --prefix)/etc/grafana-agent/config.yml` and add the configuration blocks for your specific telemetry needs. Refer to [Configure Grafana Agent]({{< relref "../configuration/" >}}) for more information.

{{% admonition type="note" %}}
To send your data to Grafana Cloud, set up Grafana Agent using the Grafana Cloud integration. Refer to [how to install an integration](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations/) and [macOS integration](/docs/grafana-cloud/data-configuration/integrations/integration-reference/integration-macos-node/).
{{%/admonition %}}

## Starting Grafana Agent

Open a terminal and run the following command to start Grafana Agent:

```shell
brew services start grafana-agent
```

## Viewing Grafana Agent Logs

By default, logs are written to the following locations:

* `$(brew --prefix)/var/log/grafana-agent.log`
* `$(brew --prefix)/var/log/grafana-agent.err.log`

## Upgrading Grafana Agent

Open a terminal and run the following commands to upgrade and restart Grafana Agent:

```shell
brew upgrade grafana-agent
brew services restart grafana-agent
 ```
