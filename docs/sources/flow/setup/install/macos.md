---
description: Learn how to install Grafana Agent in flow mode on macOS
title: Install Grafana Agent in flow mode on macOS
menuTitle: macOS
weight: 400
aliases:
 - ../../install/macos/
---

# Install Grafana Agent in flow mode on macOS

You can install Grafana Agent in flow mode with Homebrew on macOS.

{{% admonition type="note" %}}
The default prefix for Homebrew on Intel is `/usr/local`. The default prefix for Homebrew on Apple Silicon is `/opt/Homebrew`. You can verify the default prefix for Homebrew on your computer by opening a terminal and typing `brew --prefix`.
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

## Uninstall

To install Grafana Agent on macOS, run the following command in a terminal window:

```shell
brew uninstall grafana-agent-flow
```

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent#macos" >}})
- [Configure Grafana Agent]({{< relref "../configure/configure-macos" >}})
