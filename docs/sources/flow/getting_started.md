---
aliases:
- getting-started/
title: Getting started
weight: 200
---

# Getting Started

## Install Grafana Agent

To use Grafana Agent Flow, first [install Grafana Agent][]. Grafana Agent Flow
is an operating mode which will be available in an upcoming Grafana Agent
release.

## Running Grafana Agent Flow

Grafana Agent Flow can be enabled by setting the `EXPERIMENTAL_ENABLE_FLOW`
environment variable to `true`.

Then, use the `agent run` command to start Grafana Agent Flow, replacing
`FILE_PATH` with the path of a config file to use:

```
EXPERIMENTAL_ENABLE_FLOW=1 agent run FILE_PATH
```

> Grafana Agent Flow uses a different command-line interface and command line
> flags than the normal Grafana Agent. You can see the supported commands and
> the flags they support in the reference documentation for the [command-line
> interface][].

[command-line interface]: {{< relref "./reference/cli/" >}}

You can use this file as an example to get started:

```river
prometheus.scrape "default" {
  targets = [
    {"__address__" = "demo.robustperception.io:9090"},
  ]
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  // No endpoints configured; metrics will be accumulated locally in a WAL
  // and discarded.
}
```

[install Grafana Agent]: {{< relref "../set-up" >}}
