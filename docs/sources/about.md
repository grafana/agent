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
[Static mode]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static"
[Static mode]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/static"
[Static mode Kubernetes operator]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/operator"
[Static mode Kubernetes operator]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/operator"
[Flow mode]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow"
[Flow mode]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow"
[Prometheus]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/collect-prometheus-metrics.md"
[Prometheus]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/collect-prometheus-metrics.md"
[OTel]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/collect-opentelemetry-data.md"
[OTel]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/collect-opentelemetry-data.md"
[Loki]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/migrate/from-promtail.md"
[Loki]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/migrate/from-promtail.md"
[clustering]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/clustering/_index.md"
[clustering]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/clustering/_index.md"
[rules]: "/docs/agent/ -> /docs/agent/latest/flow/reference/components/mimir.rules.kubernetes.md"
[rules]: "/docs/grafana-cloud/ -> /docs/agent/latest/flow/reference/components/mimir.rules.kubernetes.md"
[vault]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/remote.vault.md"
[vault]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/remote.vault.md"
{{% /docs/reference %}}

[Pyroscope]: https://grafana.com/docs/pyroscope/latest/configure-client/grafana-agent/go_pull
[helm chart]: https://grafana.com/docs/grafana-cloud/monitor-infrastructure/kubernetes-monitoring/configuration/config-k8s-helmchart
[sla]: https://grafana.com/legal/grafana-cloud-sla
[observability]: https://grafana.com/docs/grafana-cloud/monitor-applications/application-observability/setup#send-telemetry

## Stability

| Project | Stability |
| ------- | --------- |
| Static mode | Stable |
| Static mode Kubernetes operator | Beta |
| Flow mode | Stable |

## Choose which variant of Grafana Agent to run

> **NOTE**: You don't have to pick just one variant; it's possible to
> mix-and-match installations of Grafana Agent.

### Compare variants

Each variant of Grafana Agent provides a different level of functionality. The following tables compare Grafana Agent Flow mode with Static mode, Operator, OpenTelemetry, and Prometheus.

#### Core telemetry

|              | Grafana Agent Flow mode  | Grafana Agent Static mode | Grafana Agent Operator | OpenTelemetry Collector | Prometheus Agent mode |
|--------------|--------------------------|---------------------------|------------------------|-------------------------|-----------------------|
| **Metrics**  | [Prometheus][], [OTel][] | Prometheus                | Prometheus             | OTel                    | Prometheus            |
| **Logs**     | [Loki][], [OTel][]       | Loki                      | Loki                   | OTel                    | No                    |
| **Traces**   | [OTel][]                 | OTel                      | OTel                   | OTel                    | No                    |
| **Profiles** | [Pyroscope][]            | No                        | No                     | Planned                 | No                    |

#### **OSS features**

|                          | Grafana Agent Flow mode | Grafana Agent Static mode | Grafana Agent Operator | OpenTelemetry Collector | Prometheus Agent mode |
|--------------------------|-------------------------|---------------------------|------------------------|-------------------------|-----------------------|
| **Kubernetes native**    | [Yes][helm chart]       | No                        | Yes                    | Yes                     | No                    |
| **Clustering**           | [Yes][clustering]       | No                        | No                     | No                      | No                    |
| **Prometheus rules**     | [Yes][rules]            | No                        | No                     | No                      | No                    |
| **Native Vault support** | [Yes][vault]            | No                        | No                     | No                      | No                    |

#### Grafana Cloud solutions

|                               | Grafana Agent Flow mode | Grafana Agent Static mode | Grafana Agent Operator | OpenTelemetry Collector | Prometheus Agent mode |
|-------------------------------|-------------------------|---------------------------|------------------------|-------------------------|-----------------------|
| **Official vendor support**   | [Yes][sla]              | Yes                       | Yes                    | No                      | No                    |
| **Cloud integrations**        | Some                    | Yes                       | Some                   | No                      | No                    |
| **Kubernetes monitoring**     | [Yes][helm chart]       | Yes, custom               | Yes                    | No                      | Yes, custom           |
| **Application observability** | [Yes][observability]    | No                        | No                     | Yes                     | No                    |

### Static mode

[Static mode][] is the original variant of Grafana Agent, introduced on March 3, 2020.
Static mode is the most mature variant of Grafana Agent.

You should run Static mode when:

* **Maturity**: You need to use the most mature version of Grafana Agent.

* **Grafana Cloud integrations**: You need to use Grafana Agent with Grafana Cloud integrations.

### Static mode Kubernetes operator

{{< admonition type="note" >}}
Grafana Agent version 0.37 and newer provides Prometheus Operator compatibility in Flow mode.
You should use Grafana Agent Flow mode for all new Grafana Agent deployments.
{{< /admonition >}}

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

### BoringCrypto

[BoringCrypto](https://pkg.go.dev/crypto/internal/boring) is an **EXPERIMENTAL** feature for building Grafana Agent
binaries and images with BoringCrypto enabled. Builds and Docker images for Linux arm64/amd64 are made available.

{{% docs/reference %}}
[integrations]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/integrations"
[integrations]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/static/configuration/integrations"

[components]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components"
[components]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components"
{{% /docs/reference %}}
