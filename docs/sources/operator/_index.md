---
title: Grafana Agent Operator
weight: 500
---

# Grafana Agent Operator (Beta)

The Grafana Agent Operator is a Kubernetes operator that makes it easier to
deploy the Grafana Agent and collect telemetry data from your pods.
It is currently in **Beta**, and is subject to change at any time.

It works by watching for [Kubernetes custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
that specify how you would like to collect telemetry data from your Kubernetes
cluster and where you would like to send it. They abstract Kubernetes-specific
configuration that is more tedious to perform manually. The Grafana Agent
Operator manages corresponding Grafana Agent deployments in your cluster by
watching for changes against the custom resources.

Metric collection is based on the [Prometheus
Operator](https://github.com/prometheus-operator/prometheus-operator) and
supports the official v1 ServiceMonitor, PodMonitor, and Probe CRDs from the
project. These custom resources represent abstractions for monitoring services,
pods, and ingresses. They are especially useful for Helm users, where manually
writing a generic SD to match all your charts can be difficult (or impossible!)
or where manually writing a specific SD for each chart can be tedious.
