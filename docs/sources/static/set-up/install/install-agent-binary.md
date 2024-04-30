---
aliases:
- ../../set-up/install-agent-binary/
- ../set-up/install-agent-binary/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/install/install-agent-binary/
- /docs/grafana-cloud/send-data/agent/static/set-up/install/install-agent-binary/
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-binary/
description: Learn how to install Grafana Agent in static mode as a standalone binary
menuTitle: Standalone
title: Install Grafana Agent in static mode as a standalone binary
weight: 700
refs:
  start:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/set-up/start-agent/#standalone-binary
    - pattern: /docs/grafana-cloud/
      destination: ../start-agent/#standalone-binary
  linux:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-linux/
    - pattern: /docs/grafana-cloud/
      destination: ./install-agent-linux/
  windows:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-on-windows/
    - pattern: /docs/grafana-cloud/
      destination: ./install-agent-on-windows/
  macos:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-macos/
    - pattern: /docs/grafana-cloud/
      destination: ./install-agent-macos/
  configure:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/
    - pattern: /docs/grafana-cloud/
      destination: ../../configuration/
---

# Install Grafana Agent in static mode as a standalone binary

Grafana Agent is distributed as a standalone binary for the following operating systems and architectures:

* Linux: AMD64, ARM64, PPC64, S390X
* macOS: AMD64, (Intel),  ARM64 (Apple Silicon)
* Windows: AMD64

{{% admonition type="note" %}}
ppc64le builds are considered secondary release targets and do not have the same level of support and testing as other platforms.
{{% /admonition %}}

The binary executable will run Grafana Agent in standalone mode. If you want to run Grafana Agent as a service, refer to the installation instructions for:

* [Linux](ref:linux)
* [macOS](ref:macos)
* [Windows](ref:windows)

## Download Grafana Agent

To download the Grafana Agent as a standalone binary, perform the following steps.

1. Navigate to the current Grafana Agent [release](https://github.com/grafana/agent/releases) page.

1. Scroll down to the **Assets** section.

1. Download the `grafana-agent` zip file that matches your operating system and machine’s architecture.

1. Extract the package contents into a directory.

1. If you are installing Grafana Agent on Linux, macOS, or FreeBSD, run the following command in a terminal:

   ```shell
   chmod +x EXTRACTED_BINARY
   ```

## Next steps

* [Start Grafana Agent](ref:start)
* [Configure Grafana Agent](ref:configure)

