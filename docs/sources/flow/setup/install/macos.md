---
aliases:
- ../../install/macos/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/macos/
description: Learn how to install Grafana Agent in flow mode on macOS
menuTitle: macOS
title: Install Grafana Agent in flow mode on macOS
weight: 400
---

# Install Grafana Agent in flow mode on macOS

You can install Grafana Agent in flow mode on macOS with Homebrew .

{{% admonition type="note" %}}
The default prefix for Homebrew on Intel is `/usr/local`. The default prefix for Homebrew on Apple Silicon is `/opt/Homebrew`. To verify the default prefix for Homebrew on your computer, open a terminal window and type `brew --prefix`.
{{% /admonition %}}

## Before you begin

* Install [Homebrew][] on your computer.

[Homebrew]: https://brew.sh

## Install

To install Grafana Agent on macOS, run the following commands in a terminal window.

1. Add the Grafana Homebrew tap:

   ```shell
   brew tap grafana/grafana
   ```

1. Install Grafana Agent:

   ```shell
   brew install grafana-agent-flow
   ```

## Upgrade

To upgrade Grafana Agent on macOS, run the following commands in a terminal window.

1. Upgrade Grafana Agent:

   ```shell
   brew upgrade grafana-agent
   ```

1. Restart Grafana Agent:

   ```shell
   brew services restart grafana-agent
   ```

## Uninstall

To uninstall Grafana Agent on macOS, run the following command in a terminal window:

```shell
brew uninstall grafana-agent-flow
```

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent#macos" >}})
- [Configure Grafana Agent]({{< relref "../configure/configure-macos" >}})
