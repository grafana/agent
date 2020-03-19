# Grafana Cloud Agent

Grafana Cloud Agent is an observability data collector optimized for sending
metrics and log data to [Grafana Cloud](https://grafana.com/products/cloud/).

Users of Prometheus cloud storage vendors with a single Prometheus instance can 
struggle sending their data at massive scale (millions of active series): 
Prometheus is sometimes called a single point of failure that generally requires 
a giant machine with a lot of resources allocated to it.

The Grafana Cloud Agent uses the same code as Prometheus, but tackles these issues 
by only using the most relevant parts of Prometheus for interaction with hosted 
metrics: 

1. Service Discovery
2. Scraping
3. Write Ahead Log (WAL)
4. Remote Write

On top of these, the Grafana Cloud Agent allows for an optional host filter
mechanism, enabling users to easily shard the Agent across their cluster and
lower the memory requirements per machine.

A typical deployment of the Grafana Cloud Agent for Prometheus metrics can see
up to a 40% reduction in memory usage with comparable scrape loads.

Despite called the "Grafana Cloud Agent," it can be utilized with any Prometheus
`remote_write` API.

## Trade-offs

By heavily optimizing Prometheus for remote write and resource reduction, some
trade-offs have been made:

- You can't query the Agent; you can only query metrics from the remote write
  storage.
- Recording rules aren't supported.
- Alerts aren't supported.
- When sharding the Agent, if your node has problems that interrupt metric
  availability, metrics tracking that node won't be sent for alerting on.

While the Agent can't use recording rules and alerts, `remote_write` systems such 
as Cortex currently support server-side rules and alerts. Note that this trade-off 
means that reliability of alerts are tied to the reliability of the remote system 
and alerts will be delayed at least by the time it takes for samples to reach 
the remote system. 

## Roadmap

- [x] Prometheus metrics
- [ ] Promtail for Loki logs
- [ ] `carbon-relay-ng` for Graphite metrics.
- [ ] A second clustering mode to solve sharding monitoring availability problems.

## Getting Started

The easiest way to get started with the Grafana Cloud Agent is to use the
Kubernetes install script. Simply copy and paste the following line in your
terminal:

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/master/production/kubernetes/install.sh)" | kubectl apply -f -
```

Other installation methods can be found in our
[Production](./production/README.md) documentation.

More detailed [documentation](./docs/README.md) is provided as part of the
repository.

## Example

A docker-compose config is provided in `example/`. It deploys the Agent, Cortex,
Grafana, and Avalanche for load testing. See the
[README in `example/`](./example/README.md) for more information.

## Getting Help

If you have any questions or feedback regarding the Grafana Cloud Agent:

* Ask a question on the Agent Slack channel. To invite yourself to the Grafana
  Slack, visit https://slack.grafana.com/ and join the #agent channel.
* [File an issue](https://github.com/grafana/agent/issues/new) for bugs, issues
  and feature suggestions.
