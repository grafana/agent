---
aliases:
- ../set-up/
- ../
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/
menuTitle: Install static mode
title: Install Grafana Agent in static mode
weight: 100
---

# Install Grafana Agent in static mode

You can install Grafana Agent in static mode on Docker, Kubernetes, Linux, macOS, or Windows.

The following architectures are supported:

- Linux: AMD64, ARM64, ARMv6, ARMv7
- Windows: AMD64
- macOS: AMD64 (Intel), ARM64 (Apple Silicon)
- FreeBSD: AMD64

{{% admonition type="note" %}}
ppc64le builds are considered secondary release targets and do not have the same level of support and testing as other platforms.
{{% /admonition %}}

{{< section >}}

{{% admonition type="note" %}}
Installing Grafana Agent on other operating systems is possible, but is not recommended or supported.
{{% /admonition %}}

## Grafana Cloud

Use the Grafana Agent [Kubernetes quickstarts](https://grafana.com/docs/grafana-cloud/kubernetes/agent-k8s/) or follow instructions for installing the Grafana Agent in the [Walkthrough](https://grafana.com/docs/grafana-cloud/quickstart/agent_linuxnode/).

## Tanka

For more information, refer to the [Tanka](https://tanka.dev) configurations in the Grafana Agent [production](https://github.com/grafana/agent/tree/main/production/tanka/grafana-agent) directory on GitHub.
