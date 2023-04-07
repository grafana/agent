---
title: Static mode
weight: 200
---

# Static mode

Static mode is the original runtime mode of Grafana Agent. Static mode is
composed of different _subsystems_:

* The _metrics subsystem_ wraps around Prometheus for collecting Prometheus
  metrics and forarding them over the Prometheus `remote_write` protocol.

* The _logs subsystem_ wraps around Grafana Promtail for collecting logs and
  forwarding them to Grafana Loki.

* The _traces subsystem_ wraps around OpenTelemetry Collector for collecting
  traces and forwarding them to Grafana Tempo or any OpenTelemetry-compatible
  endpoint.
