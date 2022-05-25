<p align="center"><img src="docs/user/assets/logo_and_name.png" alt="Grafana Agent logo"></p>

Grafana Agent is a telemetry collector for sending metrics, logs,
and trace data to the opinionated Grafana observability stack. It works best
with:

* [Grafana Cloud](https://grafana.com/products/cloud/)
* [Grafana Enterprise Stack](https://grafana.com/products/enterprise/)
* OSS deployments of [Grafana Loki](https://grafana.com/oss/loki/), [Prometheus](https://prometheus.io/), [Grafana Mimir](https://grafana.com/oss/mimir/), and [Grafana Tempo](https://grafana.com/oss/tempo/)

Users of Prometheus operating at a massive scale (i.e., millions of active
series) can struggle to run an unsharded singleton Prometheus instance: it becomes a
single point of failure and requires a giant machine with a lot of resources
allocated to it. Even with proper sharding across multiple Prometheus instances,
using Prometheus to send data to a cloud vendor can seem redundant: why pay for
cloud storage if data is already stored locally?

The Grafana Agent uses the same code as Prometheus, but tackles these issues
by only using the most relevant parts of Prometheus for interaction with hosted
metrics:

1. Service Discovery
2. Scraping
3. Write Ahead Log (WAL)
4. Remote Write

On top of these, the Grafana Agent enables easier sharding mechanisms that
enable users to shard Agents across their cluster and lower the memory requirements
per machine.

A typical deployment of the Grafana Agent for Prometheus metrics can see
up to a 40% reduction in memory usage with equal scrape loads.

The Grafana Agent it can be used to send Prometheus metrics to any system that
supports the Prometheus `remote_write` API.

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
as Mimir currently support server-side rules and alerts. Note that this trade-off
means that reliability of alerts are tied to the reliability of the remote system
and alerts will be delayed at least by the time it takes for samples to reach
the remote system.

## Getting Started

When using Kubernetes this [link](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s) offers the best guide.

Other installation methods can be found in our
[Grafana Agent](./docs/user/set-up/_index.md) documentation.

More detailed [documentation](./docs/README.md) is provided as part of the
repository.

## Example

The [`example/`](./example) folder contains docker-compose configs and a local
k3d/Tanka environment. Both examples deploy the Agent, Cortex and Grafana for
testing the agent. See the [docker-compose README](./example/docker-compose/README.md)
and the [k3d example README](./example/k3d/README.md) for more information.

## Prometheus Vendoring

The Grafana Agent vendors a downstream Prometheus repository maintained by
[Grafana Labs](https://github.com/grafana/prometheus). This is done so
experimental features Grafana Labs wants to contribute upstream can first be
tested and iterated on quickly within the Agent. We aim to keep the
experimental changes to a minimum and upstream changes as soon as possible.

For more context on our vendoring strategy, read our
[downstream repo maintenance guide](./docs/developer/downstream-prometheus.md).

## Getting Help

If you have any questions or feedback regarding the Grafana Agent:

* Ask a question on the Agent Slack channel. To invite yourself to the Grafana
  Slack, visit https://slack.grafana.com/ and join the #agent channel.
* Alternatively ask questions on the
  [Discussions page](https://github.com/grafana/agent/discussions).
* [File an issue](https://github.com/grafana/agent/issues/new) for bugs, issues
  and feature suggestions.
* Attend the [Grafana Agent Community Call](https://docs.google.com/document/d/1TqaZD1JPfNadZ4V81OCBPCG_TksDYGlNlGdMnTWUSpo).

## Contributing

Any contributions are welcome and details can be found
[in our contributors guide](./docs/developer/contributing.md).
