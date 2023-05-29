---
title: macOS
weight: 130
aliases:
- ../../set-up/install-agent-macos/
---

# Install Grafana Agent on macOS

You can use Homebrew to install Grafana Agent on macOS.

{{% admonition type="note" %}}
You can also install Grafana Agent on macOS using the macOS binary.
{{%/admonition %}}

## Before you begin

Ensure that [Homebrew][] is installed on your machine.

[Homebrew]: https://brew.sh

## Install Grafana Agent with Homebrew

1. Open a terminal and run the following commands:

   ```
   brew update
   brew install grafana-agent
   ```

    The brew install command downloads Grafana Agent and installs it at:
    - `/usr/local/Cellar/grafana-agent/VERSION` (Homebrew v2)
    - `/opt/homebrew/Cellar/grafana-agent/VERSION` (Homebrew v3)
    
    By default, Grafana Agent logs are located in `/opt/homebrew/var/log/`.

1. Open a terminal and run the following commands:

    ```
    mkdir -p $(brew --prefix)/etc/grafana-agent/
    touch $(brew --prefix)/etc/grafana-agent/config.yml
    ```

1. Modify `config.yml` with your configuration requirements.

    Refer to [Configure Grafana Agent]({{< relref "../configuration/" >}}) for informatuon about the Agent configuration .

1. Open a terminal and run the following command to start Grafana Agent:

    ` brew services start grafana-agent`

    For logs, see:
    - stdout: `$(brew --prefix)/var/log/grafana-agent.log`
    - stderr: `$(brew --prefix)/var/log/grafana-agent.err.log`

1. Open a terminal and run the following command to upgrade Grafana Agent:

    `brew upgrade grafana-agent`.

{{% admonition type="note" %}}
If you want to send your data to Grafana Cloud, set up Grafana Agent using the Grafana Cloud integration. Refer to [how to install an integration](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations/) and [macOS integration](/docs/grafana-cloud/data-configuration/integrations/integration-reference/integration-macos-node/).
{{%/admonition %}}
