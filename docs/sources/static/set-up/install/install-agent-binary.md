---
title: Install Grafana Agent in static mode as a standalone binary
menuTitle: Standalone
weight: 700
aliases:
- ../../set-up/install-agent-binary/
- ../install-agent-binary/
---

# Install Grafana Agent in static mode

Grafana Agent is distributed as a standalone binary for various operating systems and architectures:

* Linux: AMD64, ARM64, PPC64, S390X
* macOS: AMD64, (Intel),  ARM64 (Apple Silicon)
* Windows: AMD64

The binary executable will run Grafana Agent in standalone mode. If you want to run Grafana Agent as a service, refer to the installation instructions for:

* [Linux]({{< relref "./install-agent-linux.md" >}})
* [macOS]({{< relref "./install-agent-macos.md" >}})
* [Windows]({{< relref "./install-agent-on-windows.md" >}})

ppc64le builds are considered secondary release targets and do not have the same level of support and testing as other platforms.

## Download Grafana Agent

To download the Grafana Agent as a standalone binary, perform the following steps.

1. Navigate to the current Grafana Agent [release](https://github.com/grafana/agent/releases).

1. Scroll down to the **Assets** section.

1. Download the `grafana-agent` version that matches your operating system and machine’s architecture.

1. Extract the package contents into a directory.

1. If you are installing Grafana Agent on Linux, macOS, or FreeBSD, run the following command in a terminal:

   ```shell
   chmod +x EXTRACTED_BINARY
   ```

## Configure Grafana Agent

Refer to [Create a configuration file]({{< relref "../../configuration/create-config-file/" >}}) for information about editing or creating a configuration file.

## Run Grafana Agent

1. Open a terminal on Linux or macOS, or open a command prompt on Windows.

1. Run the following command to start Grafana Agent in static mode:

   ```shell
   EXTRACTED_BINARY -config.file CONFIG_FILE 
   ```

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent/" >}})
- [Configure Grafana Agent]({{< relref "../../configuration/" >}})
