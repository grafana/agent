---
description: Learn how to install Grafana Agent in flow mode as a standalone binary
title: Install Grafana Agent in flow mode as a standalone binary
menuTitle: Standalone
weight: 600
aliases:
 - ../../install/binary/
---

# Install Grafana Agent in flow mode as a standalone binary

Grafana Agent is distributed as a standalone binary for the following operating systems and architectures:

* Linux: AMD64, ARM64, ARMv6, ARMv7
* Windows: AMD64
* macOS: AMD64 (Intel), ARM64 (Apple Silicon)
* FreeBSD: AMD64

## Download Grafana Agent

To download the Grafana Agent as a standalone binary, perform the following steps.

1. Navigate to the current Grafana Agent [release](https://github.com/grafana/agent/releases) page.

1. Scroll down to the **Assets** section.

1. Download the `grafana-agent` zip file that matches your operating system and machine's architecture.

1. Extract the package contents into a directory.

1. If you are installing Grafana Agent on Linux, macOS, or FreeBSD, run the following command in a terminal:

   ```shell
   chmod +x BINARY_PATH
   ```

   Replace `BINARY_PATH` with the path to the extracted binary

## Next steps

* [Start Grafana Agent]({{< relref "../start-agent#standalone-binary" >}})
* [Configure Grafana Agent]({{< relref "../configure/" >}})
