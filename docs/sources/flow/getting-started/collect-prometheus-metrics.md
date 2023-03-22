---
title: Collect Prometheus metrics
weight: 200
---

# Collect Prometheus metrics

Grafana Agent Flow can be configured to collect [Prometheus][] metrics and
forward them to any Prometheus-compatible database.

This topic describes how to:

* Configure metrics delivery

[Prometheus]: https://prometheus.io

## Components used in this topic

* [prometheus.remote_write][]
* [prometheus.scrape][]

[prometheus.remote_write]: {{< relref "../reference/components/prometheus.remote_write.md" >}}
[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md" >}}

## Before you begin

* Ensure that you basic familiarity with instrumenting applications with
  Prometheus.
* Have a set of Prometheus exports or applications exposing Prometheus metrics
  that you want to collect metrics from.
* Identify where you will write collected metrics, such as Prometheus, Grafana
  Mimir, Grafana Cloud, or Grafana Enterprise Metrics.
* Be familiar with the concept of [Components][] in Grafana Agent Flow.

[Components]: {{< relref "../concepts/components.md" >}}

## Configure metrics delivery

The [prometheus.remote_write][] component is responsible for delivering
Prometheus metrics to one or Prometheus-compatible endpoints. Once a
`prometheus.remote_write` component is defined, other Grafana Agent Flow
components can be used to forward metrics to it.

To configure a `prometheus.remote_write` component for metrics delivery,
complete the following steps:

1. Add the following `prometheus.remote_write` component to your configuration file:

   ```river
   prometheus.remote_write "LABEL" {
     endpoint {
       url = "PROMETHEUS_URL"
     }
   }
   ```

2. Replace `LABEL` with a label to use for the component, such as `default`.
   The label chosen must be unique across all `prometheus.remote_write`
   components in the same configuration file.

3. Replace `PROMETHEUS_URL` with the full URL of the Prometheus-compatible
   endpoint where metrics will be sent, such as
   `https://prometheus-us-central1.grafana.net/api/prom/push`.

4. If your endpoint requires basic authentication, complete the following:

    1. Paste the following inside of the `endpoint` block:

       ```river
       basic_auth {
         username      = "USERNAME"
         password_file = "PASSWORD_FILE"
       }
       ```

    2. Replace `USERNAME` with the basic authentication username to use.

    3. Replace `PASSWORD_FILE` with a path to a file containing the basic
       authentication password or API key.

5. If you have more than one endpoint to write metrics to, repeat the
   `endpoint` block for additional endpoints.

The following example demonstrates configuring `prometheus.remote_write` with
multiple endpoints and mixed usage of basic authentication, and a
`prometheus.scrape` component which forwards metrics to it:

```river
prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9090/api/prom/push"
  }

  endpoint {
    url = "https://prometheus-us-central1.grafana.net/api/prom/push"

    basic_auth {
      username      = env("REMOTE_WRITE_USERNAME")
      password_file = "/etc/secrets/api-key"
    }
  }
}

prometheus.scrape "example" {
  targets = [{
    __address__ = "my-application:80",
  }]

  forward_to = [prometheus.remote_write.default.receiver]
}
```

For more information on configuring metrics delivery, refer to
[prometheus.remote_write][].
