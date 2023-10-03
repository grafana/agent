---
aliases:
- ./about-agent/
- /docs/grafana-cloud/agent/about/
- /docs/grafana-cloud/monitor-infrastructure/agent/about/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/about/
canonical: https://grafana.com/docs/agent/latest/about/
menuTitle: Introduction
title: Introduction to Grafana Agent
description: Grafana Agent is a flexible, performant, vendor-neutral, telemetry collector
weight: 100
---

# Introduction to Grafana Agent

Grafana Agent is a vendor-neutral, batteries-included telemetry collector. It
is designed to be flexible, performant, and compatible with multiple ecosystems
such as Prometheus and OpenTelemetry.

Grafana Agent is available in three different variants:

- [Static mode][]: The default, original variant of Grafana Agent.
- [Static mode Kubernetes operator][]: Variant which manages agents running in Static mode.
- [Flow mode][]: The newer, more flexible re-imagining variant of Grafana Agent.

{{% docs/reference %}}
[Static mode]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/static"
[Static mode]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/static"

[Static mode Kubernetes operator]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/operator"
[Static mode Kubernetes operator]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/operator"

[Flow mode]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/flow"
[Flow mode]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT VERSION>/flow"
{{% /docs/reference %}}

## Stability

| Project | Stability |
| ------- | --------- |
| Static mode | Stable |
| Static mode Kubernetes operator | Beta |
| Flow mode | Stable |

## Choose which variant of Grafana Agent to run

> **NOTE**: You do not have to pick just one variant; it is possible to
> mix-and-match installations of Grafana Agent.

### Static mode

[Static mode][] is the original variant of Grafana Agent, first introduced on
March 3, 2020. Static mode is the most mature variant of Grafana Agent.

You should run Static mode when:

* **Maturity**: You need to use the most mature version of Grafana Agent.

* **Grafana Cloud integrations**: You need to use Grafana Agent with Grafana Cloud integrations.

* **Complete list of integrations**: You need to use a Static mode [integration][integrations] which is not yet
  available as a [component][components] in Flow mode.

### Static mode Kubernetes operator

The [Static mode Kubernetes operator][] is a variant of Grafana Agent first
introduced on June 17, 2021. It is currently in beta.

The Static mode Kubernetes operator was introduced for compatibility with
Prometheus Operator, allowing static mode to support resources from Prometheus
Operator, such as ServiceMonitors, PodMonitors, and Probes.

You should run the Static mode Kubernetes operator when:

* **Prometheus Operator compatibility**: You need to be able to consume
  ServiceMonitors, PodMonitors, and Probes from the Prometheus Operator project
  for collecting Prometheus metrics.

### Flow mode

[Flow mode][] is a stable variant of Grafana Agent first introduced on
September 29, 2022.

Flow mode was introduced as a re-imagining of Grafana Agent with a focus on
vendor neutrality, ease-of-use, improved debuggability, and ability to adapt to
the needs of power users by adopting a configuration-as-code model.

Flow mode is considered to be the future of the Grafana Agent project.
Eventually, all functionality of Static mode and the Static mode Kubernetes
operator will be added into Flow mode.

You should run Flow mode when:

* You need functionality unique to Flow mode:

  * **Debuggability**: You need to more easily debug configuration issues using
    a UI.

  * **Full OpenTelemetry support**: Support for collecting OpenTelemetry
    metrics, logs, and traces.

  * **PrometheusRule support**: Support for the PrometheusRule resource from
    the Prometheus Operator project for configuring Grafana Mimir.

  * **Ecosystem transformation**: You need to be able to convert Prometheus and
    Loki pipelines to and from OpenTelmetry Collector pipelines.

  * **Grafana Pyroscope support**: Support for collecting profiles for Grafana
    Pyroscope.


### BoringCrypto

[BoringCrypto](https://pkg.go.dev/crypto/internal/boring) is an **EXPERIMENTAL** feature for building Grafana Agent
binaries and images with BoringCrypto enabled. Builds and Docker images for Linux arm64/amd64 are made available.

{{% docs/reference %}}
[integrations]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/static/configuration/integrations"
[integrations]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations"

[components]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/flow/reference/components"
[components]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT VERSION>/flow/reference/components"
{{% /docs/reference %}}
