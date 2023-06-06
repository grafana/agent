---
title: Other systems
weight: 140
aliases:
- ../../set-up/install-agent-binary/
---

# Install Grafana Agent on other systems

Grafana Agent Flow is distributed as plain binaries for various systems:

* Linux: AMD64, ARM64, PPC64, S390X
* Windows: AMD64
* macOS: AMD64, ARM64

## Overview
Binary executables are provided for the most common operating systems. Choose the binary from the Assets list on the Releases page that matches your operating system.

ppc64le builds are considered secondary release targets and do not have the same level of support and testing as other platforms.

## Download Grafana Agent

1. Navigate to the current Grafana Agent [releases](https://github.com/grafana/agent/releases).
1. Scroll down to the **Assets** section.
1. Download the version that matches your operating system and machine’s architecture.
1. Extract the package contents into a directory.
1. If you are running Linux or macOS run the following command in a terminal:

   ```shell
   chmod +x EXTRACTED_BINARY
   ```

## Configure Grafana Agent

Refer to [Create a configuration file]({{< relref "../configuration/create-config-file/" >}}) for information about editing or creating a configuration file.

## Run Grafana Agent

1. If you are running Linux or macOS run the following command in a terminal:

   ```shell
   EXTRACTED_BINARY -config.file CONFIG_FILE
   ```

TODO: When you run in Linux... is this just stand-alone and not a systemd service? How can  you set up Linux to launch on boot (like a systemd service)? What about macOS.. anything different here?

1. If you are running Windows, 

TODO: how do you run Agent in Windows if you download the binary vs installer.