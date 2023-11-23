---
aliases:
- /docs/grafana-cloud/agent/flow/setup/install/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/install/
- /docs/grafana-cloud/send-data/agent/flow/setup/install/
- /docs/sources/flow/install/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/
description: Learn how to install Grafana Agent in flow mode
menuTitle: Install flow mode
title: Install Grafana Agent in flow mode
weight: 50
---

# Install Grafana Agent in flow mode

You can install Grafana Agent in flow mode on Docker, Kubernetes, Linux, macOS, or Windows.

The following architectures are supported:

- Linux: AMD64, ARM64
- Windows: AMD64
- macOS: AMD64 (Intel), ARM64 (Apple Silicon)
- FreeBSD: AMD64

{{% admonition type="note" %}}
Installing Grafana Agent on other operating systems is possible, but is not recommended or supported.
{{% /admonition %}}

{{< section >}}

## Data collection

By default, Grafana Agent sends anonymous usage information to Grafana Labs. Refer to [data collection][] for more information
about what data is collected and how you can opt-out.

{{% docs/reference %}}
[data collection]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/data-collection.md"
[data collection]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/data-collection.md"
{{% /docs/reference %}}
