# Grafana Cloud Agent

Grafana Cloud Agent is an observability data collector optimized for sending
metrics and log data to [Grafana Cloud](https://grafana.com/products/cloud/).

Users of Prometheus cloud storage vendors can struggle sending their data at
scale: Prometheus is a single point of failure, a "special snowflake," and
generally requires a giant machine with a lot of resources allocated to it.

The Grafana Cloud Agent tackles these issues by stripping Prometheus down to its
most relavant parts:

1. Service Discovery
2. Scraping
3. Write Ahead Log (WAL)
4. Remote Write

On top of these, the Grafana Cloud Agent allows for an optional host filter
mechanism, enabling users to distribute the resource requirements of metrics
collection by running one agent per machine.

A typical deployment of the Grafana Cloud Agent for Prometheus metrics can see
up to a 40% reduction in memory usage with comparable scrape loads.

## Trade-offs

By heavily optimizing Prometheus for remote write and resource reduction, some
trade-offs have been made:

- You can't query the Agent; you can only query metrics from the remote write
  storage.
- Recording rules aren't supported.
- Alerts aren't supported.

The Agent sets the expectation that recording rules and alerts should be the
responsibility of the remote write system rather than the responsibility of the
metrics collector.

## Roadmap

- [x] Prometheus metrics
- [ ] Promtail for Loki logs
- [ ] `carbon-relay-ng` for Graphite metrics.

## Getting Started

TODO

## Example

A docker-compose config is provided in `example/`. It deploys the Agent, Cortex,
Grafana, and Avalanche for load testing. See the
[README in example/](./example/README.md) for more information.

