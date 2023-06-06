---
title: Other systems
weight: 140
aliases:
- ../../set-up/install-agent-binary/
---

# Install Grafana Agent static mode on other systems

Grafana Agent is distributed as plain binaries for the following systems:

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
1. If you are running Linux or macOS, run the following command in a terminal to make the extracted file executable:

   ```shell
   chmod +x EXTRACTED_BINARY
   ```

## Configure Grafana Agent

Refer to [Create a configuration file]({{< relref "../configuration/create-config-file/" >}}) for information about editing or creating a configuration file.

## Run Grafana Agent

1. Open a terminal on Linux or macOS, or  a command prompt on Windows.
1. Run the following command to start Grafana Agent:

   ```shell
   EXTRACTED_BINARY -config.file CONFIG_FILE
   ```
