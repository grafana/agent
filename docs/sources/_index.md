---
title: Grafana Agent
weight: 1
---

# Grafana Agent

Grafana Agent is software that can run on any of your hosts, and which collects events and metrics from hosts and sends them to Grafana.  Grafana Agent is capable of gathering data for [all three pillars of observability](https://grafana.com/docs/grafana-cloud/fundamentals/intro-to-observability/); logs, metrics, and traces.

Grafana Agent can collect, transform, and send data to:

* The [Prometheus][] ecosystem
* The [OpenTelemetry][] ecosystem
* The Grafana open source ecosystem ([Loki][], [Grafana][], [Tempo][], [Mimir][], [Phlare][])

[Terraform]: https://terraform.io
[Grafana Agent Flow]: {{< relref "./flow/" >}}
[About Grafana Agent]: {{< relref "./about.md" >}}
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

Grafana Agent is based around **components**. Components are wired together to
form programmable observability **pipelines** for telemetry collection,
processing, and delivery.

> **NOTE**: This page focuses mainly on "[Flow mode][Grafana Agent Flow]," the
> Terraform-inspired variant of Grafana Agent.
>
> For information on other variants of Grafana Agent, refer to [About Grafana
> Agent][].

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
