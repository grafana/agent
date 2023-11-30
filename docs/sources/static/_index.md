---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/static/
- /docs/grafana-cloud/send-data/agent/static/
canonical: https://grafana.com/docs/agent/latest/static/
description: Learn about Grafana Agent in static mode
title: Static mode
weight: 200
---

# Static mode

Static mode is the original mode of Grafana Agent, and is the most mature.
Static mode is composed of different _subsystems_:

* The _metrics subsystem_ wraps around Prometheus for collecting Prometheus
  metrics and forwarding them over the Prometheus `remote_write` protocol.

* The _logs subsystem_ wraps around Grafana Promtail for collecting logs and
  forwarding them to Grafana Loki.

* The _traces subsystem_ wraps around OpenTelemetry Collector for collecting
  traces and forwarding them to Grafana Tempo or any OpenTelemetry-compatible
  endpoint.

Static mode is [configured][configure] with a YAML file.

Static mode works with:

- Grafana Cloud
- Grafana Enterprise Stack
- OSS deployments of Grafana Loki, Grafana Mimir, Grafana Tempo, and Prometheus

This topic helps you to think about what you're trying to accomplish and how to
use Grafana Agent to meet your goals.

You can [set up][] and [configure][] Grafana Agent in static mode manually, or
you can follow the common workflows described in this topic.

## Topics

### Static mode Grafana Agent for Grafana Cloud integrations

There are different ways for you to set up Grafana Agent to scrape
data&mdash;through Grafana's integration platform or directly. Select a guide
to get started:

| Topic | Description |
|---|---|
| [Get started with monitoring using an integration](/docs/grafana-cloud/data-configuration/get-started-integration/) | Walk through installing a Linux integration using Grafana Agent in the Grafana Cloud interface. |
| [Install and manage integrations](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations/)  | View general steps for using Grafana Cloud integrations to install Grafana Agent to collect data. See [supported integrations](/docs/grafana-cloud/data-configuration/integrations/integration-reference/).
| [Ship your metrics to Grafana Cloud without an integration](/docs/grafana-cloud/data-configuration/metrics/agent-config-exporter/) | If you want to ship your Prometheus metrics to Grafana Cloud but there isn’t an integration available, you can use a Prometheus exporter and deploy Grafana Agent to scrape your local machine or service. |
| [Change your metrics scrape interval](/docs/grafana-cloud/billing-and-usage/control-prometheus-metrics-usage/changing-scrape-interval/) | Learn about reducing your total data points per minute (DPM) by adjusting your scrape interval. |

### Static mode Grafana Agent for Kubernetes Monitoring

Grafana Kubernetes Monitoring provides a simplified approach to monitoring your Kubernetes fleet by deploying Grafana Agent with useful defaults for collecting metrics. Select a guide to get started monitoring Kubernetes:

| Topic | Description |
|---|---|
| [Configure Kubernetes Monitoring using Agent](/docs/grafana-cloud/kubernetes-monitoring/configuration/) | Use the Kubernetes Monitoring solution to set up monitoring of your Kubernetes cluster and to install preconfigured dashboards and alerts. |
| [Ship Kubernetes traces using Grafana Agent directly](/docs/grafana-cloud/kubernetes-monitoring/other-methods/k8s-agent-traces/) | Deploy Grafana Agent into your Kubernetes cluster as a deployment and configure it to collect traces for your Kubernetes workloads.  |

### Use Grafana Agent directly to scrape telemetry data

Grafana Cloud integration workflows and the Kubernetes Monitoring solution are the easiest ways to get started collecting telemetry data, but sometimes you might want to use a manual approach to set your configuration options.

| Topic | Description |
|---|---|
| [Install or uninstall Grafana Agent][install] | Install or uninstall Grafana Agent. |
| [Troubleshoot Cloud Integrations installation on Linux](/docs/grafana-cloud/monitor-infrastructure/integrations/install-troubleshoot-linux/) | Troubleshoot common errors when executing the Grafana Agent installation script on Linux.  |
| [Troubleshoot Cloud Integrations installation on Mac](/docs/grafana-cloud/monitor-infrastructure/integrations/install-troubleshoot-mac/) | Troubleshoot common errors when executing the Grafana Agent installation script on Mac.  |
| [Troubleshoot Cloud Integrations installation on Windows](/docs/grafana-cloud/monitor-infrastructure/integrations/install-troubleshooting-windows/) | Troubleshoot common errors when executing the Grafana Agent installation script on Windows.  |

### Use Grafana Agent to send logs to Grafana Loki

Logs are included when you [set up a Cloud integration](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations) but you can take a more hands-on approach with the following guide.

| Topic | Description |
|---|---|
| [Collect logs with Grafana Agent](/docs/grafana-cloud/data-configuration/logs/collect-logs-with-agent/) |  Install Grafana Agent to collect logs for use with Grafana Loki, included with your [Grafana Cloud account](/docs/grafana-cloud/account-management/cloud-portal/). |

### Use Grafana Agent to send traces to Grafana Tempo

| Topic | Description |
|---|---|
| [Set up and use tracing](/docs/grafana-cloud/data-configuration/traces/set-up-and-use-tempo/) |  Install Grafana Agent to collect traces for use with Grafana Tempo, included with your [Grafana Cloud account](/docs/grafana-cloud/account-management/cloud-portal/). |
| [Use Grafana Agent as a tracing pipeline](/docs/tempo/latest/configuration/grafana-agent/) | Grafana Agent can be configured to run a set of tracing pipelines to collect data from your applications and write it to Grafana Tempo. Pipelines are built using OpenTelemetry, and consist of receivers, processors, and exporters. |

{{% docs/reference %}}
[set up]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up"
[set up]: "/docs/grafana-cloud/ -> ./set-up"
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration"
[configure]: "/docs/grafana-cloud/ -> ./configuration"
[install]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/install"
[install]: "/docs/grafana-cloud/ -> ./set-up/install"
{{% /docs/reference %}}
