---
aliases:
- /docs/grafana-cloud/agent/
- /docs/grafana-cloud/monitor-infrastructure/agent/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/
- /docs/grafana-cloud/send-data/agent/
canonical: https://grafana.com/docs/agent/latest/
title: Grafana Agent
description: Grafana Agent is a flexible, performant, vendor-neutral, telemetry collector
weight: 350
cascade:
  AGENT_RELEASE: $AGENT_VERSION
  OTEL_VERSION: v0.96.0
refs:
  variants:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/about/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/about/
  static-mode:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/
  static-mode-kubernetes-operator:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/operator/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/operator/
  flow-mode:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/
    - pattern: /docs/grafana-cloud/
      destination: /docs/agent/<AGENT_VERSION>/flow/
  ui:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/tasks/debug/#grafana-agent-flow-ui
    - pattern: /docs/grafana-cloud/
      destination: /docs/agent/<AGENT_VERSION>/flow/tasks/debug/#grafana-agent-flow-ui
---

# Grafana Agent

Grafana Agent is an OpenTelemetry Collector distribution with configuration
inspired by [Terraform][]. It is designed to be flexible, performant, and
compatible with multiple ecosystems such as Prometheus and OpenTelemetry.

Grafana Agent is based around **components**. Components are wired together to
form programmable observability **pipelines** for telemetry collection,
processing, and delivery.

{{< admonition type="note" >}}
This page focuses mainly on [Flow mode](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/), the Terraform-inspired variant of Grafana Agent.

For information on other variants of Grafana Agent, refer to [Introduction to Grafana Agent](../about/).
{{< /admonition >}}

Grafana Agent can collect, transform, and send data to:

* The [Prometheus][] ecosystem
* The [OpenTelemetry][] ecosystem
* The Grafana open source ecosystem ([Loki][], [Grafana][], [Tempo][], [Mimir][], [Pyroscope][])

[Terraform]: https://terraform.io
[Prometheus]: https://prometheus.io
[OpenTelemetry]: https://opentelemetry.io
[Loki]: https://github.com/grafana/loki
[Grafana]: https://github.com/grafana/grafana
[Tempo]: https://github.com/grafana/tempo
[Mimir]: https://github.com/grafana/mimir
[Pyroscope]: https://github.com/grafana/pyroscope

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
  [built-in UI](ref:ui).
* **Batteries included**: Integrate with systems like MySQL, Kubernetes, and
  Apache to get telemetry that's immediately useful.

## Getting started

* Choose a [variant](ref:variants) of Grafana Agent to run.
* Refer to the documentation for the variant to use:
  * [Static mode](ref:static-mode)
  * [Static mode Kubernetes operator](ref:static-mode-kubernetes-operator)
  * [Flow mode](ref:flow-mode)

## Supported platforms

The following operating systems and hardware architecture are supported.

## Linux

* Minimum version: kernel 4.x or later
* Architectures: AMD64, ARM64
* Within the Linux distribution lifecycle

## Windows

* Minimum version: Windows Server 2016 or later, or Windows 10 or later.
* Architectures: AMD64

## macOS

* Minimum version: macOS 10.13 or later
* Architectures: AMD64 on Intel, ARM64 on Apple Silicon

## FreeBSD

* Within the FreeBSD lifecycle
* Architectures: AMD64

## Release cadence

A new minor release is planned every six weeks for the entire Grafana Agent
project, including Static mode, the Static mode Kubernetes operator, and Flow
mode.

The release cadence is best-effort: if necessary, releases may be performed
outside of this cadence, or a scheduled release date can be moved forwards or
backwards.

Minor releases published on cadence include updating dependencies for upstream
OpenTelemetry Collector code if new versions are available. Minor releases
published outside of the release cadence may not include these dependency
updates.

Patch and security releases may be created at any time.

