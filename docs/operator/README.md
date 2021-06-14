# Grafana Agent Operator

The Grafana Agent Operator is a Kubernetes operator that makes it easier to
deploy the Grafana Agent and easier to collect telemetry data from your pods.

Metric collection is based on the [Prometheus
Operator](https://github.com/prometheus-operator/prometheus-operator) and
supports the official v1 ServiceMonitor, PodMonitr, and Probe CRDs from the
project.

## Table of Contents

1. [Getting Started](./getting-started.md)
  1. [Deploying CustomResourceDefinitions](./getting-started.md#deploying-customresourcedefinitions)
  2. [Installing on Kubernetes](./getting-started.md#installing-on-kubernetes)
  3. [Running locally](./getting-started.md#running-locally)
  4. [Deploying GrafanaAgent](./getting-started.md#deploying-grafanagent)
2. [FAQ](./faq.md)
3. [Architecture](./architecture.md)
4. [Maintainers Guide](./maintainers-guide.md)
