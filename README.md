<p align="center"><img src="docs/sources/assets/logo_and_name.png" alt="Grafana Agent logo"></p>

Grafana Agent is a vendor-neutral, batteries-included telemetry collector with
configuration inspired by [Terraform][]. It is designed to be flexible,
performant, and compatible with multiple ecosystems such as Prometheus and
OpenTelemetry.

Grafana Agent is based around **components**. Components are wired together to
form programmable observability **pipelines** for telemetry collection,
processing, and delivery.

> **NOTE**: This page focuses mainly on "[Flow mode][Grafana Agent Flow]," the
> Terraform-inspired revision of Grafana Agent.

Grafana Agent can collect, transform, and send data to:

* The [Prometheus][] ecosystem
* The [OpenTelemetry][] ecosystem
* The Grafana open source ecosystem ([Loki][], [Grafana][], [Tempo][], [Mimir][], [Pyroscope][])

[Terraform]: https://terraform.io
[Grafana Agent Flow]: https://grafana.com/docs/agent/latest/flow/
[Prometheus]: https://prometheus.io
[OpenTelemetry]: https://opentelemetry.io
[Loki]: https://github.com/grafana/loki
[Grafana]: https://github.com/grafana/grafana
[Tempo]: https://github.com/grafana/tempo
[Mimir]: https://github.com/grafana/mimir
[Pyroscope]: https://github.com/grafana/pyroscope

## Why use Grafana Agent?

* **Vendor-neutral**: Fully compatible with the Prometheus, OpenTelemetry, and
  Grafana open source ecosystems.
* **Every signal**: Collect telemetry data for metrics, logs, traces, and
  continuous profiles.
* **Scalable**: Deploy on any number of machines to collect millions of active
  series and terabytes of logs.
* **Battle-tested**: Grafana Agent extends the existing battle-tested code from
  the Prometheus and OpenTelemetry Collector projects.
* **Powerful**: Write programmable pipelines with ease, and debug them using a
  [built-in UI][UI].
* **Batteries included**: Integrate with systems like MySQL, Kubernetes, and
  Apache to get telemetry that's immediately useful.

[UI]: https://grafana.com/docs/agent/latest/flow/monitoring/debugging/#grafana-agent-flow-ui

## Getting started

Check out our [documentation][] to see:

* [Installation instructions][] for Grafana Agent Flow
* Details about [Grafana Agent Flow][]
* Steps for [Getting started][] with Grafana Agent Flow
* The list of Grafana Agent Flow [Components][]

[documentation]: https://grafana.com/docs/agent/latest/
[Installation instructions]: https://grafana.com/docs/agent/latest/flow/setup/install/
[Grafana Agent Flow]: https://grafana.com/docs/agent/latest/flow/
[Getting started]: https://grafana.com/docs/agent/latest/flow/getting_started/
[Components]: https://grafana.com/docs/agent/latest/flow/reference/components/

## Example

```river
// Discover Kubernetes pods to collect metrics from.
discovery.kubernetes "pods" {
  role = "pod"
}

// Collect metrics from Kubernetes pods.
prometheus.scrape "default" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [prometheus.remote_write.default.receiver]
}

// Get an API key from disk.
local.file "apikey" {
  filename  = "/var/data/my-api-key.txt"
  is_secret = true
}

// Send metrics to a Prometheus remote_write endpoint.
prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9009/api/prom/push"

    basic_auth {
      username = "MY_USERNAME"
      password = local.file.apikey.content
    }
  }
}
```

We maintain an example [Docker Compose environment][] that can be used to
launch dependencies to play with Grafana Agent locally.

[Docker Compose environment]: ./example/docker-compose/

## Release cadence

A new minor release is planned every six weeks.

The release cadence is best-effort: releases may be moved forwards or backwards
if needed. The planned release dates for future minor releases do not change if
one minor release is moved.

Patch and security releases may be created at any time.

## Community

To engage with the Grafana Agent community:

* Chat with us on our community Slack channel. To invite yourself to the
  Grafana Slack, visit <https://slack.grafana.com/> and join the `#agent`
  channel.
* Ask questions on the [Discussions page][].
* [File an issue][] for bugs, issues, and feature suggestions.
* Attend the monthly [community call][].

[Discussions page]: https://github.com/grafana/agent/discussions
[File an issue]: https://github.com/grafana/agent/issues/new
[community call]: https://docs.google.com/document/d/1TqaZD1JPfNadZ4V81OCBPCG_TksDYGlNlGdMnTWUSpo

## Contribute

Refer to our [contributors guide][] to learn how to contribute.

[contributors guide]: ./docs/developer/contributing.md
