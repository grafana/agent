---
aliases:
- ./about-agent/
- /docs/grafana-cloud/agent/about/
- /docs/grafana-cloud/monitor-infrastructure/agent/about/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/about/
- /docs/grafana-cloud/send-data/agent/about/
canonical: https://grafana.com/docs/agent/latest/about/
description: Grafana Agent is a flexible, performant, vendor-neutral, telemetry collector
menuTitle: Introduction
title: Introduction to Grafana Agent
weight: 100
---

# Introduction to Grafana Agent

Grafana Agent is a flexible, high performance, vendor-neutral telemetry collector. It's fully compatible with the most popular open source observability standards such as OpenTelemetry (OTel) and Prometheus.

Grafana Agent is available in three different variants:

- [Static mode][]: The original Grafana Agent.
- [Static mode Kubernetes operator][]: The Kubernetes operator for Static mode.
- [Flow mode][]: The new, component-based Grafana Agent.

{{% docs/reference %}}
[Static mode]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/static"
[Static mode]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/static"

[Static mode Kubernetes operator]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/operator"
[Static mode Kubernetes operator]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/operator"

[Flow mode]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/flow"
[Flow mode]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT VERSION>/flow"

[Prometheus]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/collect-prometheus-metrics.md"
[Prometheus]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/collect-prometheus-metrics.md"

[OTel]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/collect-opentelemetry-data.md"
[OTel]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/collect-opentelemetry-data.md"

[Loki]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/migrating-from-promtail.md"
[Loki]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/getting-started/migrating-from-promtail.md"
{{% /docs/reference %}}

## Stability

| Project | Stability |
| ------- | --------- |
| Static mode | Stable |
| Static mode Kubernetes operator | Beta |
| Flow mode | Stable |

## Choose which variant of Grafana Agent to run

> **NOTE**: You don't have to pick just one variant; it's possible to
> mix-and-match installations of Grafana Agent.

### Static mode

[Static mode][] is the original variant of Grafana Agent, introduced on March 3, 2020.
Static mode is the most mature variant of Grafana Agent.

You should run Static mode when:

* **Maturity**: You need to use the most mature version of Grafana Agent.

* **Grafana Cloud integrations**: You need to use Grafana Agent with Grafana Cloud integrations.

### Static mode Kubernetes operator

The [Static mode Kubernetes operator][] is a variant of Grafana Agent introduced on June 17, 2021. It's currently in beta.

The Static mode Kubernetes operator provides compatibility with Prometheus Operator,
allowing static mode to support resources from Prometheus Operator, such as ServiceMonitors, PodMonitors, and Probes.

You should run the Static mode Kubernetes operator when:

* **Prometheus Operator compatibility**: You need to be able to consume
  ServiceMonitors, PodMonitors, and Probes from the Prometheus Operator project
  for collecting Prometheus metrics.

### Flow mode

[Flow mode][] is a stable variant of Grafana Agent, introduced on September 29, 2022.

Grafana Agent Flow mode focuses on vendor neutrality, ease-of-use,
improved debugging, and ability to adapt to the needs of power users by adopting a configuration-as-code model.

You should run Flow mode when:

* You need functionality unique to Flow mode:

  * **Improved debugging**: You need to more easily debug configuration issues using a UI.

  * **Full OpenTelemetry support**: Support for collecting OpenTelemetry metrics, logs, and traces.

  * **PrometheusRule support**: Support for the PrometheusRule resource from the Prometheus Operator project for configuring Grafana Mimir.

  * **Ecosystem transformation**: You need to be able to convert Prometheus and Loki pipelines to and from OpenTelmetry Collector pipelines.

  * **Grafana Pyroscope support**: Support for collecting profiles for Grafana Pyroscope.

#### Core telemetry

|              | Prometheus Agent mode | Grafana Agent Static mode | Grafana Agent Operator | OpenTelemetry Collector | Grafana Agent Flow mode  |
|--------------|-----------------------|---------------------------|------------------------|-------------------------|--------------------------|
| **Metrics**  | Prometheus            | Prometheus                | Prometheus             | OTel                    | [Prometheus][], [OTel][] |
| **Logs**     | No                    | Loki                      | Loki                   | OTel                    | [Loki][], [OTel][]       |
| **Traces**   | No                    | OTel                      | OTel                   | OTel                    | [OTel][]                 |
| **Profiles** | No                    | No                        | No                     | No                      | Pyroscope                |

#### **OSS features**

|                          | Prometheus Agent mode | Grafana Agent Static mode | Grafana Agent Operator | OpenTelemetry Collector | Grafana Agent Flow mode |
|--------------------------|-----------------------|---------------------------|------------------------|-------------------------|-------------------------|
| **Kubernetes native**    | No                    | No                        | Yes                    | Yes                     | Yes                     |
| **Clustering**           | No                    | No                        | No                     | No                      | No                      |
| **Prometheus rules**     | No                    | No                        | No                     | No                      | Yes                     |
| **Native Vault support** | No                    | No                        | No                     | No                      | Yes                     |

#### Grafana Cloud solutions

|                               | Prometheus Agent mode | Grafana Agent Static mode | Grafana Agent Operator | OpenTelemetry Collector | Grafana Agent Flow mode |
|-------------------------------|-----------------------|---------------------------|------------------------|-------------------------|-------------------------|
| **Official vendor support**   | No                    | Yes                       | Yes                    | No                      | Yes                     |
| **Cloud integrations**        | No                    | Yes                       | Some                   | No                      | Some                    |
| **Kubernetes monitoring**     | Yes, custom           | Yes, custom               | Yes                    | No                      | Yes                     |
| **Application observability** | No                    | No                        | No                     | Yes                     | Yes                     |


### BoringCrypto

[BoringCrypto](https://pkg.go.dev/crypto/internal/boring) is an **EXPERIMENTAL** feature for building Grafana Agent
binaries and images with BoringCrypto enabled. Builds and Docker images for Linux arm64/amd64 are made available.

{{% docs/reference %}}
[integrations]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/static/configuration/integrations"
[integrations]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations"

[components]: "/docs/agent/ -> /docs/agent/<AGENT VERSION>/flow/reference/components"
[components]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT VERSION>/flow/reference/components"
{{% /docs/reference %}}
