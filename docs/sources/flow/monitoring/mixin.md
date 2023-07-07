---
title: Compile and install agent-flow-mixin
weight: 400
---

# Compile and install agent-flow-mixin

A [mixin][] is a set of preconfigured dashboards, alerts and recording rules,
packaged in a reusable and extensible bundle.

This topic describes how to compile and install the [agent-flow-mixin][] to
reuse the set of dashboards and alerts the agent team maintains as their
opinionated way of monitoring Grafana Agent Flow.

## Before you begin

* Install [mixtool][]

## Steps

To compile and install agent-flow-mixin

1. Clone the grafana/agent repo.

1. Run `make build-mixin` from the repo root.

1. Check the compiled mixin on the `operations/agent-flow-mixin-compiled`
directory.

    The mixin will also be available as a .zip package on `operations/agent-flow-mixin.zip` for easier transport between machines.

1. [Import][] the compiled JSON dashboards on
   `operations/agent-flow-mixin-compiled/dashboards` to Grafana.

1. [Upload][] the compiled alerting rules on
   `operations/agent-flow-mixin-compiled/alerts.yaml` to your Prometheus
configuration.

1. Alternatively if you're managing your infrastructure with Jsonnet, use
   [jsonnet-bundler][] and/or [grizzly][] to install the mixin directly.


[mixin]: https://grafana.com/blog/2018/09/13/everything-you-need-to-know-about-monitoring-mixins/
[agent-flow-mixin]: https://github.com/grafana/agent/tree/main/operations/agent-flow-mixin
[mixtool]: https://github.com/monitoring-mixins/mixtool
[Import]: https://grafana.com/docs/grafana/latest/dashboards/manage-dashboards/#import-a-dashboard
[Upload]: https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
[jsonnet-bundler]: https://github.com/jsonnet-bundler/jsonnet-bundler
[grizzly]: https://github.com/grafana/grizzly/
