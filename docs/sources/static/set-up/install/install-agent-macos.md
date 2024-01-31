---
aliases:
- ../../set-up/install-agent-macos/
- ../set-up/install-agent-macos/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/install/install-agent-macos/
- /docs/grafana-cloud/send-data/agent/static/set-up/install/install-agent-macos/
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-macos/
description: Learn how to install Grafana Agent in static mode on macOS
menuTitle: macOS
title: Install Grafana Agent in static mode on macOS
weight: 500
---

# Install Grafana Agent in static mode on macOS

You can install Grafana Agent in static mode on macOS with Homebrew.

## Before you begin

Install [Homebrew][] on your computer.

{{< admonition type="note" >}}
The default prefix for Homebrew on Intel is `/usr/local`. The default prefix for Homebrew on Apple Silicon is `/opt/Homebrew`. To verify the default prefix for Homebrew on your computer, open a terminal window and type `brew --prefix`.
{{< /admonition >}}

[Homebrew]: https://brew.sh

## Install

To install Grafana Agent on macOS, run the following commands in a terminal window.

1. Update Homebrew:

   ```shell
   brew update
   ```

1. Install Grafana Agent:

   ```shell
   brew install grafana-agent
   ```

Grafana Agent is installed by default at `$(brew --prefix)/Cellar/grafana-agent/VERSION`.

## Upgrade

To upgrade Grafana Agent on macOS, run the following commands in a terminal window.

1. Upgrade Grafana Agent:

   ```shell
   brew upgrade grafana-agent
   ```

1. Restart Grafana Agent:

   ```shell
   brew services restart grafana-agent

## Uninstall

To uninstall Grafana Agent on macOS, run the following command in a terminal window:

```shell
brew uninstall grafana-agent
```

## Configure

1. To create the Agent `config.yml` file, open a terminal and run the following command:

    ```shell
    touch $(brew --prefix)/etc/grafana-agent/config.yml
    ```

1. Edit `$(brew --prefix)/etc/grafana-agent/config.yml` and add the configuration blocks for your specific telemetry needs. Refer to [Configure Grafana Agent][configure] for more information.

{{< admonition type="note" >}}
To send your data to Grafana Cloud, set up Grafana Agent using the Grafana Cloud integration. Refer to [how to install an integration](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations/) and [macOS integration](/docs/grafana-cloud/data-configuration/integrations/integration-reference/integration-macos-node/).
{{< /admonition >}}

## Next steps

- [Start Grafana Agent][start]
- [Configure Grafana Agent][configure]

{{% docs/reference %}}
[start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/start-agent"
[start]: "/docs/grafana-cloud/ -> ../start-agent"
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/create-config-file"
[configure]: "/docs/grafana-cloud/ -> ../../configuration/create-config-file"
{{% /docs/reference %}}
