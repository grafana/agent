---
draft: "True"
---

<p align="center"><img src="assets/logo_and_name.png" alt="Grafana Agent logo"></p>

Grafana Agent is a telemetry collector for sending metrics, logs,
and trace data to the opinionated Grafana observability stack. It works best
with:

* [Grafana Cloud](https://grafana.com/products/cloud/)
* [Grafana Enterprise Stack](https://grafana.com/products/enterprise/)
* OSS deployments of [Grafana Loki](https://grafana.com/oss/loki/), [Prometheus](https://prometheus.io/), [Cortex](https://cortexmetrics.io/), and [Grafana Tempo](https://grafana.com/oss/tempo/)


- Grafana Agent uses less memory on average than Prometheus – by doing less
  (only focusing on `remote_write`-related functionality).
- Grafana Agent allows for deploying multiple instances of the Agent in a
  cluster and only scraping metrics from targets that are running on the same host.
  This allows for distributing memory requirements across the cluster
  rather than pressurizing a single node.
