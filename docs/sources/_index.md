---
canonical: https://grafana.com/docs/agent/latest/
title: Grafana Agent
weight: 550
---

# Grafana Agent

Grafana Agent is a vendor-neutral, batteries-included telemetry collector with
configuration inspired by [Terraform][]. It is designed to be flexible,
performant, and compatible with multiple ecosystems such as Prometheus and
OpenTelemetry.

Grafana Agent is based around **components**. Components are wired together to
form programmable observability **pipelines** for telemetry collection,
processing, and delivery.

{{% admonition type="note" %}}
This page focuses mainly on [Flow mode]({{< relref "./flow/" >}}), the Terraform-inspired variant of Grafana Agent.

For information on other variants of Grafana Agent, refer to [Introduction to Grafana Agent]({{< relref "./about.md" >}}).
{{% /admonition %}}

Grafana Agent can collect, transform, and send data to:

* The [Prometheus][] ecosystem
* The [OpenTelemetry][] ecosystem
* The Grafana open source ecosystem ([Loki][], [Grafana][], [Tempo][], [Mimir][], [Phlare][])

[Terraform]: https://terraform.io
[Prometheus]: https://prometheus.io
[OpenTelemetry]: https://opentelemetry.io
[Loki]: https://github.com/grafana/loki
[Grafana]: https://github.com/grafana/grafana
[Tempo]: https://github.com/grafana/tempo
[Mimir]: https://github.com/grafana/mimir
[Phlare]: https://github.com/grafana/phlare

## Why use Grafana Agent?

* **Vendor-neutral**: Fully compatible with the Prometheus, OpenTelemetry, and
  Grafana open source ecosystems.
* **Every signal**: Collect telemetry data for metrics, logs, traces, and
  continuous profiles.
* **Scalable**: Deploy on any number of machines to collect millions of active
  series and terabytes of logs.
* **Battle-tested**: Grafana Agent extends the existing battle-tested code from
  the Prometheus and OpenTelemetry Collector projects.
* **Powerful**: Write programmable pipelines with ease, and debug them using a
  [built-in UI][UI].
* **Batteries included**: Integrate with systems like MySQL, Kubernetes, and
  Apache to get telemetry that's immediately useful.

[UI]: {{< relref "./flow/monitoring/debugging.md#grafana-agent-flow-ui" >}}

## Getting started

* Choose a [variant][variants] of Grafana Agent to run.
* Refer to the documentation for the variant to use:
  * [Static mode][]
  * [Static mode Kubernetes operator][]
  * [Flow mode][]

[variants]: {{< relref "./about.md" >}}
[Static mode]: {{< relref "./static/" >}}
[Static mode Kubernetes operator]: {{< relref "./operator/" >}}
[Flow mode]: {{< relref "./flow/" >}}

## Supported platforms

* Linux

  * Minimum version: kernel 2.6.32 or later
  * Architectures: AMD64, ARM64

* Windows

  * Minimum version: Windows Server 2012 or later, or Windows 10 or later.
  * Architectures: AMD64

* macOS

  * Minimum version: macOS 10.13 or later
  * Architectures: AMD64 (Intel), ARM64 (Apple Silicon)

* FreeBSD

  * Minimum version: FreeBSD 10 or later
  * Architectures: AMD64

## Release cadence

A new minor release is planned every six weeks for the entire Grafana Agent
project, including Static mode, the Static mode Kubernetes operator, and Flow
mode.

The release cadence is best-effort: releases may be moved forwards or backwards
if needed. The planned release dates for future minor releases do not change if
one minor release is moved.

Patch and security releases may be created at any time.

[Milestones]: https://github.com/grafana/agent/milestones
