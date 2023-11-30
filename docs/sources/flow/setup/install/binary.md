---
aliases:
- ../../install/binary/
- /docs/grafana-cloud/agent/flow/setup/install/binary/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/binary/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/install/binary/
- /docs/grafana-cloud/send-data/agent/flow/setup/install/binary/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/binary/
description: Learn how to install Grafana Agent Flow as a standalone binary
menuTitle: Standalone
title: Install Grafana Agent  Flow as a standalone binary
weight: 600
---

# Install {{< param "PRODUCT_NAME" >}} as a standalone binary

{{< param "PRODUCT_NAME" >}} is distributed as a standalone binary for the following operating systems and architectures:

* Linux: AMD64, ARM64
* Windows: AMD64
* macOS: AMD64 (Intel), ARM64 (Apple Silicon)
* FreeBSD: AMD64

## Download {{< param "PRODUCT_ROOT_NAME" >}}

To download {{< param "PRODUCT_NAME" >}} as a standalone binary, perform the following steps.

1. Navigate to the current {{< param "PRODUCT_ROOT_NAME" >}} [release](https://github.com/grafana/agent/releases) page.

1. Scroll down to the **Assets** section.

1. Download the `grafana-agent` zip file that matches your operating system and machine's architecture.

1. Extract the package contents into a directory.

1. If you are installing {{< param "PRODUCT_NAME" >}} on Linux, macOS, or FreeBSD, run the following command in a terminal:

   ```shell
   chmod +x BINARY_PATH
   ```

   Replace `BINARY_PATH` with the path to the extracted binary

## Next steps

{{< param "PRODUCT_NAME" >}}

{{% docs/reference %}}
[Start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/setup/start-agent.md#standalone-binary"
[Start]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/setup/start-agent.md#standalone-binary"
[Configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/setup/configure"
[Configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/setup/configure/"
{{% /docs/reference %}}
