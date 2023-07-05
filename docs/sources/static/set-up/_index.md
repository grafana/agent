---
title: Install Grafana Agent in static mode
menuTitle: Install static mode
weight: 100
aliases:
- ../set-up/
---

# Install Grafana Agent in static mode

You can install Grafana Agent in static mode on Docker, Kubernetes, Linux, macOS, or Windows.

{{% admonition type="note" %}}
Installing Grafana Agent on other operating systems is possible, but is not recommended or supported.
{{% /admonition %}}

The following architectures are supported:

- Linux: AMD64, ARM64, ARMv6, ARMv7
- Windows: AMD64
- macOS: AMD64 (Intel), ARM64 (Apple Silicon)
- FreeBSD: AMD64

In addition, best-effort support is provided for Linux: ppc64le.

To get started with Grafana Agent Operator, refer to the Operator-specific
[documentation](../../operator/_index.md).

{{< section >}}

### Grafana Cloud

Use the Grafana Agent [Kubernetes quickstarts](https://grafana.com/docs/grafana-cloud/kubernetes/agent-k8s/) or follow instructions for installing the Grafana Agent in the [Walkthrough](https://grafana.com/docs/grafana-cloud/quickstart/agent_linuxnode/).

### Tanka

For more information, refer to the [Tanka](https://tanka.dev) configurations in our [`production/`](https://github.com/grafana/agent/tree/main/production/tanka/grafana-agent) directory.
