# Grafana Agent Operator

The Grafana Agent Operator is a Kubernetes operator that makes it easier to
deploy Grafana Agent and easier to discover targets for metric collection.

It is based on the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
and aims to be compatible the official ServiceMonitor, PodMonitor, and Probe
CRDs that Prometheus Operator users are used to.

## Roadmap

- [ ] Helm chart
- [ ] Logs support
- [ ] Traces support
- [ ] Integrations support

## Documentation

Refer to the project's [documentation](../../docs/operator) for how to install
and get started with the Grafana Agent Operator.

## Developer Reference

The [Maintainer's Guide](../../docs/operator/maintainers-guide.md) includes
basic information to help you understand how the code works. This can be very
useful if you are planning on working on the operator.
