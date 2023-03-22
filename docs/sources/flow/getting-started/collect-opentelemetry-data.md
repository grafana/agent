---
title: Collect OpenTelemetry data
weight: 300
---

# Collect OpenTelemetry data

Grafana Agent Flow can be configured to collect [OpenTelemetry][] data and
forward it to any OpenTelemetry-compatible endpoint.

This topic describes how to:

* Configure OpenTelemetry data delivery

[OpenTelemetry]: https://opentelemetry.io

## Components used in this topic

* [otelcol.auth.basic][]
* [otelcol.exporter.otlp][]
* [otelcol.exporter.otlphttp][]

[otelcol.auth.basic]: {{< relref "../reference/components/otelcol.auth.basic.md" >}}
[otelcol.exporter.otlp]: {{< relref "../reference/components/otelcol.exporter.otlp.md" >}}
[otelcol.exporter.otlphttp]: {{< relref "../reference/components/otelcol.exporter.otlphttp.md" >}}

## Before you begin

* Ensure that you have basic familiarity with instrumenting applications with
  OpenTelemetry.
* Have a set of OpenTelemetry applications ready to push telemetry data to
  Grafana Agent Flow.
* Identify where Grafana Agent Flow will write received telemetry data.
* Be familiar with the concept of [Components][] in Grafana Agent Flow.

[Components]: {{< relref "../concepts/components.md" >}}

## Configure an OpenTelemetry exporter

Before components can receive OpenTelemetry data, you must have a component
responsible for exporting the OpenTelemetry data. An OpenTelemetry _exporter
component_ is responsible for writing (that is, exporting) OpenTelemetry data
to an external system.

In this task, we will use the [otelcol.exporter.otlp][] component to send
OpenTelemetry data to a server using the OpenTelemetry Protocol (OTLP). Once an
exporter component is defined, other Grafana Agent Flow components can be used
to forward data to it.

> Refer to the list of available [Components][] for the full list of
> `otelcol.exporter` components that can be used to export OpenTelemetry data.
>
> [Components]: {{< relref "../reference/components/" >}}

To configure an `otelcol.exporter.otlp` component for exporting OpenTelemetry
data, complete the following steps:

1. Add the following `otelcol.exporter.otlp` component to your configuration
   file:

   ```river
   otelcol.exporter.otlp "LABEL" {
     client {
       url = "HOST:PORT"
     }
   }
   ```

    1. Replace `LABEL` with a label to use for the component, such as `default`.
       The label chosen must be unique across all `otelcol.exporter.otlp`
       components in the same configuration file.

    2. Replace `HOST` with the hostname or IP address of the server to send
       OpenTelemetry data to.

    3. Replace `PORT` with the port of the server to send OpenTelemetry data to.

2. If your server requires basic authentication, complete the following:

    1. Add the following `otelcol.auth.basic` component to your configuration file:

       ```river
       otelcol.auth.basic "BASIC_AUTH_LABEL" {
         username = "USERNAME"
         password = "PASSWORD"
       }
       ```

        1. Replace `BASIC_AUTH_LABEL` with a label to use for the component, such
           as `default`. The label chosen must be unique across all
           `otelcol.auth.basic` components in the same configuration file.

        2. Replace `USERNAME` with the basic authentication username to use.

        3. Replace `PASSWORD` with the basic authentication password or API key to
           use.

    2. Add the following line inside of the `client` block of your
       `otelcol.exporter.otlp` component:

       ```
       auth = otelcol.auth.basic.BASIC_AUTH_LABEL.handler
       ```

        1. Replace `BASIC_AUTH_LABEL` with the label used for the
           `otelcol.auth.basic` component in step 2.1.1.

3. If you have more than one server to export metrics to, create a new
   `otelcol.exporter.otlp` component for each additional server.

> `otelcol.exporter.otlp` sends data using OTLP over gRPC (HTTP/2). To send to
> a server using HTTP/1.1, follow the steps above but use the
> [otelcol.exporter.otlphttp component][otelcol.exporter.otlphttp] instead.

The following example demonstrates configuring `otelcol.exporter.otlp` with
authentication and a component which forwards data to it:

```river
otelcol.exporter.otlp "default" {
  client {
    endpoint = "my-otlp-grpc-server:4317"
    auth     = otelcol.auth.basic.credentials.handler
  }
}

otelcol.auth.basic "credentials" {
  // Retrieve credentials using environment variables.

  username = env("BASIC_AUTH_USER")
  password = env("API_KEY")
}

otelcol.receiver.otlp "example" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.exporter.otlp.default.input]
    logs    = [otelcol.exporter.otlp.default.input]
    traces  = [otelcol.exporter.otlp.default.input]
  }
}
```

For more information on configuring OpenTelemetry data delivery using OTLP,
refer to [otelcol.exporter.otlp][].
