---
title: Mixin
weight: 400
---

# Mixin

A [mixin][] is a set of preconfigured dashboards, alerts and recording rules,
packaged in a reusable and extensible bundle. The agent team maintains the
[agent-flow-mixin][] as an opinionated way of monitoring Grafana Agent Flow
deployments.

Use the mixin by installing [mixtool][] and calling `make build-mixin`
form the grafana/agent repo root.

The compiled mixin is available on the `operations/agent-flow-mixin-compiled`
directory and also as a zip package on `operations/agent-flow-mixin.zip`.

The dashboards and alerts can then just be imported into Grafana and Prometheus
respectively.

```
$ go install github.com/monitoring-mixins/mixtool/cmd/mixtool@main
$ git clone https://github.com/grafana/agent.git
$ cd agent
$ make build-mixin
$ tree operations/agent-flow-mixin-compiled
operations/agent-flow-mixin-compiled
├── alerts.yaml
└── dashboards
    ├── agent-cluster-node.json
    ├── agent-cluster-overview.json
    ├── agent-flow-controller.json
    ├── agent-flow-prometheus.remote_write.json
    └── agent-flow-resources.json
```

## Clustering

The mixin contains a set of predefined dashboards and alerts for monitoring
[agent clusters][], allowing to both get an overview of the current state, as
well as easily drill down to node-level information with the click of a button.

![](../../../assets/clustering_overview_dashboard.png)
![](../../../assets/clustering_node_info_dashboard.png)
![](../../../assets/clustering_node_transport_dashboard.png)

[mixin]: https://grafana.com/blog/2018/09/13/everything-you-need-to-know-about-monitoring-mixins/
[agent-flow-mixin]: https://github.com/grafana/agent/tree/main/operations/agent-flow-mixin
[mixtool]: https://github.com/monitoring-mixins/mixtool
[agent clusters]: {{< relref "../concepts/clustering.md" >}}
