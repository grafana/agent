---
canonical: https://grafana.com/docs/agent/latest/flow/tasks/migrate/from-otelcol/
description: Learn how to migrate from OpenTelemetry Collector to Grafana Agent Flow
menuTitle: Migrate from OpenTelemetry Collector
title: Migrate from OpenTelemetry Collector to Grafana Agent Flow
weight: 310
---

# Migrate from OpenTelemetry Collector to {{% param "PRODUCT_NAME" %}}

The built-in {{< param "PRODUCT_ROOT_NAME" >}} convert command can migrate your [OpenTelemetry Collector][] configuration to a {{< param "PRODUCT_NAME" >}} configuration.

This topic describes how to:

* Convert an OpenTelemetry Collector configuration to a {{< param "PRODUCT_NAME" >}} configuration.
* Run an OpenTelemetry Collector configuration natively using {{< param "PRODUCT_NAME" >}}.

## Components used in this topic

* [otelcol.receiver.otlp][]
* [otelcol.processor.memory_limiter][]
* [otelcol.exporter.otlp][]

## Before you begin

* You must have an existing OpenTelemetry Collector configuration.
* You must have a set of OpenTelemetry Collector applications ready to push telemetry data to {{< param "PRODUCT_NAME" >}}.
* You must be familiar with the concept of [Components][] in {{< param "PRODUCT_NAME" >}}.

## Convert an OpenTelemetry Collector configuration

To fully migrate your configuration from [OpenTelemetry Collector] to {{< param "PRODUCT_NAME" >}}, you must convert your OpenTelemetry Collector configuration into a {{< param "PRODUCT_NAME" >}} configuration.
This conversion will enable you to take full advantage of the many additional features available in {{< param "PRODUCT_NAME" >}}.

> In this task, you will use the [convert][] CLI command to output a {{< param "PRODUCT_NAME" >}}
> configuration from a OpenTelemetry Collector configuration.

1. Open a terminal window and run the following command.

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=otelcol --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=otelcol --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
   ```

   {{< /code >}}

   Replace the following:

   - _`<INPUT_CONFIG_PATH>`_: The full path to the OpenTelemetry Collector configuration.
   - _`<OUTPUT_CONFIG_PATH>`_: The full path to output the {{< param "PRODUCT_NAME" >}} configuration.

1. [Run][] {{< param "PRODUCT_NAME" >}} using the new {{< param "PRODUCT_NAME" >}} configuration from _`<OUTPUT_CONFIG_PATH>`_:

### Debugging

1. If the `convert` command can't convert an OpenTelemetry Collector configuration, diagnostic information is sent to `stderr`.\
   You can bypass any non-critical issues and output the {{< param "PRODUCT_NAME" >}} configuration using a best-effort conversion by including the `--bypass-errors` flag.

    {{< admonition type="caution" >}}
    If you bypass the errors, the behavior of the converted configuration may not match the original OpenTelemetry Collector configuration.
    Make sure you fully test the converted configuration before using it in a production environment.
    {{< /admonition >}}

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=otelcol --bypass-errors --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=otelcol --bypass-errors --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
   ```

   {{< /code >}}

   Replace the following:

   - _`<INPUT_CONFIG_PATH>`_: The full path to the OpenTelemetry Collector configuration.
   - _`<OUTPUT_CONFIG_PATH>`_: The full path to output the {{< param "PRODUCT_NAME" >}} configuration.

1. You can also output a diagnostic report by including the `--report` flag.

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=otelcol --report=<OUTPUT_REPORT_PATH> --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=otelcol --report=<OUTPUT_REPORT_PATH> --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
   ```

   {{< /code >}}

   Replace the following:

   - _`<INPUT_CONFIG_PATH>`_: The full path to the OpenTelemetry Collector configuration.
   - _`<OUTPUT_CONFIG_PATH>`_: The full path to output the {{< param "PRODUCT_NAME" >}} configuration.
   - _`<OUTPUT_REPORT_PATH>`_: The output path for the report.

    Using the [example][] OpenTelemetry Collector configuration below, the diagnostic report provides the following information:

    ```plaintext
    (Info) Converted receiver/otlp into otelcol.receiver.otlp.default
    (Info) Converted processor/memory_limiter into otelcol.processor.memory_limiter.default
    (Info) Converted exporter/otlp into otelcol.exporter.otlp.default

    A configuration file was generated successfully.
    ```

## Run an OpenTelemetry Collector configuration

If you’re not ready to completely switch to a {{< param "PRODUCT_NAME" >}} configuration, you can run {{< param "PRODUCT_ROOT_NAME" >}} using your existing OpenTelemetry Collector configuration.
The `--config.format=otelcol` flag tells {{< param "PRODUCT_ROOT_NAME" >}} to convert your OpenTelemetry Collector configuration to a {{< param "PRODUCT_NAME" >}} configuration and load it directly without saving the new configuration.
This allows you to try {{< param "PRODUCT_NAME" >}} without modifying your existing OpenTelemetry Collector configuration infrastructure.

> In this task, you will use the [run][] CLI command to run {{< param "PRODUCT_NAME" >}}
> using an OpenTelemetry Collector configuration.

[Run][] {{< param "PRODUCT_NAME" >}} and include the command line flag `--config.format=otelcol`.
Your configuration file must be a valid OpenTelemetry Collector configuration file rather than a {{< param "PRODUCT_NAME" >}} configuration file.

### Debugging

1. You can follow the convert CLI command [debugging][] instructions to generate a diagnostic report.

1. Refer to the {{< param "PRODUCT_NAME" >}} [Debugging][DebuggingUI] for more information about a running {{< param "PRODUCT_NAME" >}}.

1. If your OpenTelemetry Collector configuration can't be converted and loaded directly into {{< param "PRODUCT_NAME" >}}, diagnostic information is sent to `stderr`.
   You can bypass any non-critical issues and start the Agent by including the `--config.bypass-conversion-errors` flag in addition to `--config.format=otelcol`.

   {{< admonition type="caution" >}}
   If you bypass the errors, the behavior of the converted configuration may not match the original Prometheus configuration.
   Do not use this flag in a production environment.
   {{< /admonition >}}

## Example

This example demonstrates converting an OpenTelemetry Collector configuration file to a {{< param "PRODUCT_NAME" >}} configuration file.

The following OpenTelemetry Collector configuration file provides the input for the conversion.

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

exporters:
  otlp:
    endpoint: database:4317

processors:
  memory_limiter:
    limit_percentage: 90
    check_interval: 1s


service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [memory_limiter]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [memory_limiter]
      exporters: [otlp]
    traces:
      receivers: [otlp]
      processors: [memory_limiter]
      exporters: [otlp]
```

The convert command takes the YAML file as input and outputs a [River][] file.

{{< code >}}

```static-binary
AGENT_MODE=flow grafana-agent convert --source-format=otelcol --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
```

```flow-binary
grafana-agent-flow convert --source-format=otelcol --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
```

{{< /code >}}

Replace the following:

- _`<INPUT_CONFIG_PATH>`_: The full path to the OpenTelemetry Collector configuration.
- _`<OUTPUT_CONFIG_PATH>`_: The full path to output the {{< param "PRODUCT_NAME" >}} configuration.

The new {{< param "PRODUCT_NAME" >}} configuration file looks like this:

```river
otelcol.receiver.otlp "default" {
	grpc { }

	http { }

	output {
		metrics = [otelcol.processor.memory_limiter.default.input]
		logs    = [otelcol.processor.memory_limiter.default.input]
		traces  = [otelcol.processor.memory_limiter.default.input]
	}
}

otelcol.processor.memory_limiter "default" {
	check_interval   = "1s"
	limit_percentage = 90

	output {
		metrics = [otelcol.exporter.otlp.default.input]
		logs    = [otelcol.exporter.otlp.default.input]
		traces  = [otelcol.exporter.otlp.default.input]
	}
}

otelcol.exporter.otlp "default" {
	client {
		endpoint = "database:4317"
	}
}
```

## Limitations

Configuration conversion is done on a best-effort basis. {{< param "PRODUCT_ROOT_NAME" >}} will issue warnings or errors where the conversion can't be performed.

After the configuration is converted, review the {{< param "PRODUCT_NAME" >}} configuration file created and verify that it's correct before starting to use it in a production environment.

The following list is specific to the convert command and not {{< param "PRODUCT_NAME" >}}:

* Components are supported which directly embed upstream OpenTelemetry Collector features. You can get a general idea of which exist in
  {{< param "PRODUCT_NAME" >}} for conversion by reviewing the `otelcol.*` components in the [Component Reference][].
  Any additional unsupported features are returned as errors during conversion.
* Check if you are using any extra command line arguments with OpenTelemetry Collector that aren't present in your configuration file.
* Metamonitoring metrics exposed by {{< param "PRODUCT_NAME" >}} usually match OpenTelemetry Collector metamonitoring metrics but will use a different name.
  Make sure that you use the new metric names, for example, in your alerts and dashboards queries.
* The logs produced by {{< param "PRODUCT_NAME" >}} differ from those produced by OpenTelemetry Collector.
* {{< param "PRODUCT_ROOT_NAME" >}} exposes the {{< param "PRODUCT_NAME" >}} [UI][].

[OpenTelemetry Collector]: https://opentelemetry.io/docs/collector/configuration/
[debugging]: #debugging
[example]: #example

{{% docs/reference %}}
[otelcol.receiver.otlp]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.otlp.md"
[otelcol.processor.memory_limiter]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.memory_limiter.md"
[otelcol.exporter.otlp]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.otlp.md"
[Components]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/components.md"
[Components]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/concepts/components.md"
[Component Reference]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/_index.md"
[Component Reference]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components/_index.md"
[convert]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/cli/convert.md"
[convert]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/cli/convert.md"
[run]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/cli/run.md"
[run]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/cli/run.md"
[Run]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/get-started/run/"
[Run]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/get-started/run/"
[DebuggingUI]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/debug.md"
[DebuggingUI]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/debug.md"
[River]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/config-language/_index.md"
[River]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/concepts/config-language/_index.md"
[UI]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/debug#grafana-agent-flow-ui"
[UI]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/debug#grafana-agent-flow-ui"
{{% /docs/reference %}}
