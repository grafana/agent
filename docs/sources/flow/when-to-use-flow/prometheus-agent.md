---
description: Learn how to Flow compares to the Prometheus Agent Mode
menuTitle: Prometheus Agent Mode
title: Prometheus Agent Mode
weight: 200
---

## Should you use Flow or the Prometheus Agent Mode?

Grafana Agent Flow supports all of the [features][prom-agent-mode-blog] of the [Prometheus Agent Mode][prom-agent-mode-flag] via components such as:
* `discovery.http`
* `discovery.relabel`
* `prometheus.scrape`
* `prometheus.relabel`
* `prometheus.remote_write`

Agent Flow's performance is comparable to Prometheus, because `prometheus` Flow components are abe 
to processes Prometheus metrics natively without converting them to other formats such as OTLP.

Agent Flow also comes pre-built with exporters out of the box. For example, `prometheus.exporter.unix` 
provides the functionality of Prometheus' [Node Exporter][node-exporter].

[prom-agent-mode-flag]: https://prometheus.io/docs/prometheus/latest/feature_flags/#prometheus-agent
[prom-agent-mode-blog]: https://prometheus.io/blog/2021/11/16/agent/
[node-exporter]: https://github.com/prometheus/node_exporter