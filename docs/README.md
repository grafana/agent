# Grafana Agent Documentation

This directory contains documentation for Grafana Agent. It is split into two
parts:

* `user/`: Documentation for users. This directory is hosted on
  [grafana.com/docs/agent](https://grafana.com/docs/agent/latest/) and we
  recommend interacting with it there instead of viewing the markdown on
  GitHub.
* `developer/`: Documentation for contributors and maintainers.
* `rfcs/`: RFCs for proposals relating to Grafana Agent.

## Preview the website

Run `make docs`. This launches a preview of the website with the current grafana docs at `http://localhost:3002/docs/agent/latest/` which will refresh automatically when changes are made to content in the `sources` directory.
Make sure Docker is running.

### Community Projects

Below is a list of community-led projects for working with Grafana Agent. These projects are not maintained or supported by Grafana Labs.

#### Helm (Kubernetes Deployment)

A publicly available release of a Grafana Agent Helm chart is maintained [here](https://github.com/DandyDeveloper/charts/tree/master/charts/grafana-agent). Contributions and improvements are welcomed. Full details on rolling out and supported options can be found in the [readme](https://github.com/DandyDeveloper/charts/blob/master/charts/grafana-agent/README.md).

This *does not* require the Grafana Agent Operator to rollout / deploy.

#### Juju (Charmed Operator)

The [grafana-agent-k8s](https://github.com/canonical/grafana-agent-operator) charmed operator runs with [Juju](https://juju.is) the Grafana Agent on Kubernetes.
The Grafana Agent charmed operator is designed to work with the [Logs, Metrics and Alerts](https://juju.is/docs/lma2) observability stack.
