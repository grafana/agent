---
title: Collect OpenTelemetry data
weight: 300
---

# Collect OpenTelemetry data

Grafana Agent Flow can be configured to collect [OpenTelemetry][] data and
forward it to any OpenTelemetry-compatible endpoint.

This topic describes how to:

* Configure OpenTelemetry data delivery
* Configure batching
* Receive OpenTelemetry data over OTLP

[OpenTelemetry]: https://opentelemetry.io

## Components used in this topic

* [otelcol.auth.basic][]
* [otelcol.exporter.otlp][]
* [otelcol.exporter.otlphttp][]
* [otelcol.processor.batch][]
* [otelcol.receiver.otlp][]

[otelcol.auth.basic]: {{< relref "../reference/components/otelcol.auth.basic.md" >}}
[otelcol.exporter.otlp]: {{< relref "../reference/components/otelcol.exporter.otlp.md" >}}
[otelcol.exporter.otlphttp]: {{< relref "../reference/components/otelcol.exporter.otlphttp.md" >}}
[otelcol.processor.batch]: {{< relref "../reference/components/otelcol.processor.batch.md" >}}
[otelcol.receiver.otlp]: {{< relref "../reference/components/otelcol.receiver.otlp.md" >}}

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
   otelcol.exporter.otlp "EXPORTER_LABEL" {
     client {
       url = "HOST:PORT"
     }
   }
   ```

    1. Replace `EXPORTER_LABEL` with a label to use for the component, such as
       `default`. The label chosen must be unique across all
       `otelcol.exporter.otlp` components in the same configuration file.

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

## Configure batching

Production-ready Grafana Agent Flow components will not send OpenTelemetry data
directly to an exporter for delivery. Instead, data is usually sent to one or
more _processor components_ that perform various transformations on the data.

Ensuring data is batched is a production-readiness step to improve the
compression of data and reduce the number of outgoing network requests to
external systems.

In this task, we will configure an [otelcol.processor.batch][] component to
batch data before sending it to our exporter.

> Refer to the list of available [Components][] for the full list of
> `otelcol.processor` components that can be used to process OpenTelemetry
> data. You can chain processors by having one processor send data to another
> processor.
>
> [Components]: {{< relref "../reference/components/" >}}

To configure an `otelcol.processor.batch` component, complete the following
steps:

1. Follow [Configure an OpenTelemetry
   exporter](#configure-an-opentelemetry-exporter) to ensure received data can
   be written to an external system.

2. Add the following `otelcol.processor.batch` component into your
   configuration file:

   ```river
   otelcol.processor.batch "LABEL" {
     output {
       metrics = [otelcol.exporter.otlp.EXPORTER_LABEL.input]
       logs    = [otelcol.exporter.otlp.EXPORTER_LABEL.input]
       traces  = [otelcol.exporter.otlp.EXPORTER_LABEL.input]
     }
   }
   ```

    1. Replace `LABEL` with a label to use for the component, such as
       `default`. The label chosen must be unique across all
       `otelcol.processor.batch` components in the same configuration file.

    2. Replace `EXPORTER_LABEL` with the label for your existing
       `otelcol.exporter.otlp` component.

    3. To disable one of the telemetry types, set the relevant type in the
       `output` block to the empty list, such as `metrics = []`.

    4. To send batched data to another processor, replace the components in the
       `output` list with the processor components to use.

The following example demonstrates configuring a sequence of
`otelcol.processor` components before ultimately being exported:

```river
otelcol.processor.memory_limiter "default" {
  check_interval = "1s"
  limit          = "1GiB"

  output {
    metrics = [otelcol.processor.batch.default.input]
    logs    = [otelcol.processor.batch.default.input]
    traces  = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
    logs    = [otelcol.exporter.otlp.default.input]
    traces  = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = "my-otlp-grpc-server:4317"
  }
}
```

For more information on configuring OpenTelemetry data batching, refer to
[otelcol.processor.batch][].

## Configure an OpenTelemetry receiver

Grafana Agent Flow can be configured to receive OpenTelemetry metrics, logs,
and traces. An OpenTelemetry _receiver_ component is responsible for receiving
OpenTelemetry data from an external system.

In this task, we will use the [otelcol.receiver.otlp][] component to receive
OpenTelemetry data over the network using the OpenTelemetry Protocol (OTLP). A
receiver component can be configured to forward received data to other Grafana
Agent Flow components.

> Refer to the list of available [Components][] for the full list of
> `otelcol.receiver` components that can be used to receive OpenTelemetry data.
>
> [Components]: {{< relref "../reference/components/" >}}

To configure an `otelcol.receiver.otlp` component for receiving OpenTelemetry
data, complete the following steps:

1. Follow [Configure an OpenTelemetry
   exporter](#configure-an-opentelemetry-exporter) to ensure received data can
   be written to an external system.

2. Optional: Follow [Configure batching](#configure-batching) to improve
   compression and reduce the total amount of network requests.

3. Add the following `otelcol.receiver.otlp` component to your configuration
   file:

   ```river
   otelcol.receiver.otlp "LABEL" {
     output {
       metrics = [COMPONENT_INPUT_LIST]
       logs    = [COMPONENT_INPUT_LIST]
       traces  = [COMPONENT_INPUT_LIST]
     }
   }
   ```

    1. Replace `LABEL` with a label to use for the component, such as
       `default`. The label chosen must be unique across all
       `otelcol.receiver.otlp` components in the same configuration file.

    2. Replace `COMPONENT_INPUT_LIST` with a comma-delimited list of component
       inputs to forward received data to. For example, to send data to an
       existing batch processor component, use
       `otelcol.processor.batch.PROCESSOR_LABEL.input`. To send data directly
       to an existing exporter component, use
       `otelcol.exporter.otlp.EXPORTER_LABEL.input`.

    3. To enable receiving OTLP data over gRPC on port `4317`, add `grpc {}` to
       your `otelcol.receiver.otlp` component.

    4. To enable receiving OTLP data over HTTP on port `4318`, add `http {}` to
       your `otelcol.receiver.otlp` component.

    5. To disable one of the telemetry types, set the relevant type in the
       `output` block to the empty list, such as `metrics = []`.

The following example demonstrates configuring `otelcol.receiver.otlp` and
sending it to an exporter:

```river
otelcol.receiver.otlp "example" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.processor.batch.example.input]
    logs    = [otelcol.processor.batch.example.input]
    traces  = [otelcol.processor.batch.example.input]
  }
}

otelcol.processor.batch "example" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
    logs    = [otelcol.exporter.otlp.default.input]
    traces  = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = "my-otlp-grpc-server:4317"
  }
}
```

For more information on configuring OpenTelemetry data delivery using OTLP,
refer to [otelcol.receiver.otlp][].
