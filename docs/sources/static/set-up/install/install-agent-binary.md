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
---

# Install Grafana Agent in static mode as a standalone binary

Grafana Agent is distributed as a standalone binary for the following operating systems and architectures:

* Linux: AMD64, ARM64, PPC64, S390X
* macOS: AMD64, (Intel),  ARM64 (Apple Silicon)
* Windows: AMD64

{{< admonition type="note" >}}
ppc64le builds are considered secondary release targets and do not have the same level of support and testing as other platforms.
{{< /admonition >}}

The binary executable will run Grafana Agent in standalone mode. If you want to run Grafana Agent as a service, refer to the installation instructions for:

* [Linux][linux]
* [macOS][macos]
* [Windows][windows]

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

* [Start Grafana Agent][start]
* [Configure Grafana Agent][configure]

{{% docs/reference %}}
[linux]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-linux"
[linux]: "/docs/grafana-cloud/ -> ./install-agent-linux"
[macos]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-macos"
[macos]: "/docs/grafana-cloud/ -> ./install-agent-macos"
[windows]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/install/install-agent-on-windows"
[windows]: "/docs/grafana-cloud/ -> ./install-agent-on-windows"
[start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/start-agent#standalone-binary"
[start]: "/docs/grafana-cloud/ -> ../start-agent#standalone-binary"
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration"
[configure]: "/docs/grafana-cloud/ -> ../../configuration"
{{% /docs/reference %}}
