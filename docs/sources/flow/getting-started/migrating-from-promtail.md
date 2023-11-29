---
aliases:
- /docs/grafana-cloud/agent/flow/getting-started/migrating-from-promtail/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/migrating-from-promtail/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/getting-started/migrating-from-promtail/
- /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-promtail/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/migrating-from-promtail/
description: Learn how to migrate from Promtail to Grafana Agent Flow
menuTitle: Migrate from Promtail
title: Migrate from Promtail to Grafana Agent Flow
weight: 330
---

# Migrate from Promtail to {{< param "PRODUCT_NAME" >}}

The built-in {{< param "PRODUCT_ROOT_NAME" >}} convert command can migrate your [Promtail][]
configuration to a {{< param "PRODUCT_NAME" >}} configuration.

This topic describes how to:

* Convert a Promtail configuration to a {{< param "PRODUCT_NAME" >}} configuration.
* Run a Promtail configuration natively using {{< param "PRODUCT_NAME" >}}.

## Components used in this topic

* [local.file_match][]
* [loki.source.file][]
* [loki.write][]

## Before you begin

* You must have an existing Promtail configuration.
* You must be familiar with the concept of [Components][] in {{< param "PRODUCT_NAME" >}}.

## Convert a Promtail configuration

To fully migrate from [Promtail] to {{< param "PRODUCT_NAME" >}}, you must convert
your Promtail configuration into a {{< param "PRODUCT_NAME" >}} configuration. This
conversion will enable you to take full advantage of the many additional
features available in {{< param "PRODUCT_NAME" >}}.

> In this task, we will use the [convert][] CLI command to output a flow
> configuration from a Promtail configuration.

1. Open a terminal window and run the following command:

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=promtail --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=promtail --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   {{< /code >}}


   Replace the following:
    * `INPUT_CONFIG_PATH`: The full path to the Promtail configuration.
    * `OUTPUT_CONFIG_PATH`: The full path to output the flow configuration.

1. [Start][] {{< param "PRODUCT_NAME" >}} using the new flow configuration
   from `OUTPUT_CONFIG_PATH`:

### Debugging

1. If the convert command cannot convert a Promtail configuration, diagnostic
   information is sent to `stderr`. You can bypass any non-critical issues and
   output the flow configuration using a best-effort conversion by including
   the `--bypass-errors` flag.

   {{% admonition type="caution" %}}
   If you bypass the errors, the behavior of the converted configuration may not match the original Promtail configuration.
   Make sure you fully test the converted configuration before using it in a production environment.
   {{% /admonition %}}

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=promtail --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=promtail --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   {{< /code >}}

1. You can also output a diagnostic report by including the `--report` flag.

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=promtail --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=promtail --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   {{< /code >}}

    * Replace `OUTPUT_REPORT_PATH` with the output path for the report.

   Using the [example](#example) Promtail configuration below, the diagnostic
   report provides the following information:

    ```plaintext
    (Warning) If you have a tracing set up for Promtail, it cannot be migrated to {{< param "PRODUCT_NAME" >}} automatically. Refer to the documentation on how to configure tracing in {{< param "PRODUCT_NAME" >}}.
    (Warning) The metrics from {{< param "PRODUCT_NAME" >}} are different from the metrics emitted by Promtail. If you rely on Promtail's metrics, you must change your configuration, for example, your alerts and dashboards.
    ```

## Run a Promtail configuration

If you’re not ready to completely switch to a flow configuration, you can run
{{< param "PRODUCT_ROOT_NAME" >}} using your existing Promtail configuration.
The `--config.format=promtail` flag tells {{< param "PRODUCT_ROOT_NAME" >}} to convert your Promtail
configuration to {{< param "PRODUCT_NAME" >}} and load it directly without saving the new
configuration. This allows you to try {{< param "PRODUCT_NAME" >}} without modifying your existing
Promtail configuration infrastructure.

> In this task, we will use the [run][] CLI command to run {{< param "PRODUCT_NAME" >}} using a Promtail configuration.

[Start][] {{< param "PRODUCT_NAME" >}} and include the command line flag
`--config.format=promtail`. Your configuration file must be a valid Promtail
configuration file rather than a Flow mode configuration file.

### Debugging

1. You can follow the convert CLI command [debugging][] instructions to generate
   a diagnostic report.

1. Refer to the {{< param "PRODUCT_NAME" >}}  [Debugging][DebuggingUI] for more information about
   running {{< param "PRODUCT_NAME" >}}.

1. If your Promtail configuration can't be converted and loaded directly into
   {{< param "PRODUCT_ROOT_NAME" >}}, diagnostic information is sent to `stderr`. You can bypass any
   non-critical issues and start {{< param "PRODUCT_ROOT_NAME" >}} by including the
   `--config.bypass-conversion-errors` flag in addition to
   `--config.format=promtail`.

   {{% admonition type="caution" %}}
   If you bypass the errors, the behavior of the converted configuration may not match the original Promtail configuration.
   Do not use this flag in a production environment.
   {{%/admonition %}}

## Example

This example demonstrates converting a Promtail configuration file to a {{< param "PRODUCT_NAME" >}} configuration file.

The following Promtail configuration file provides the input for the conversion:

```yaml
clients:
  - url: http://localhost/loki/api/v1/push
scrape_configs:
  - job_name: example
    static_configs:
      - targets:
          - localhost
        labels:
          __path__: /var/log/*.log
```

The convert command takes the YAML file as input and outputs a [River][] file.

{{< code >}}

```static-binary
AGENT_MODE=flow grafana-agent convert --source-format=promtail --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

```flow-binary
grafana-agent-flow convert --source-format=promtail --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

{{< /code >}}

The new {{< param "PRODUCT_NAME" >}} configuration file looks like this:

```river
local.file_match "example" {
	path_targets = [{
		__address__ = "localhost",
		__path__    = "/var/log/*.log",
	}]
}

loki.source.file "example" {
	targets    = local.file_match.example.targets
	forward_to = [loki.write.default.receiver]
}

loki.write "default" {
	endpoint {
		url = "http://localhost/loki/api/v1/push"
	}
	external_labels = {}
}
```

## Limitations

Configuration conversion is done on a best-effort basis. {{< param "PRODUCT_ROOT_NAME" >}} will issue
warnings or errors where the conversion can't be performed.

Once the configuration is converted, we recommend that you review
the {{< param "PRODUCT_NAME" >}} configuration file created, and verify that it's correct
before starting to use it in a production environment.

Furthermore, we recommend that you review the following checklist:

* Check if you are using any extra command line arguments with Promtail which
  aren't present in your configuration file. For example, `-max-line-size`.
* Check if you are setting any environment variables,
  whether [expanded in the config file][] itself or consumed directly by
  Promtail, such as `JAEGER_AGENT_HOST`.
* In {{< param "PRODUCT_NAME" >}}, the positions file is saved at a different location.
  Refer to the [loki.source.file][] documentation for more details. Check if you have any existing
  setup, for example, a Kubernetes Persistent Volume, that you must update to use the new
  positions file path.
* Metamonitoring metrics exposed by the Flow Mode usually match Promtail
  metamonitoring metrics but will use a different name. Make sure that you
  use the new metric names, for example, in your alerts and dashboards queries.
* Note that the logs produced by {{< param "PRODUCT_NAME" >}} will differ from those
  produced by Promtail.
* Note that {{< param "PRODUCT_NAME" >}} exposes the {{< param "PRODUCT_NAME" >}} [UI][], which differs
  from Promtail's Web UI.

[Promtail]: https://www.grafana.com/docs/loki/<LOKI_VERSION>/clients/promtail/
[debugging]: #debugging
[expanded in the config file]: https://www.grafana.com/docs/loki/<LOKI_VERSION>/clients/promtail/configuration/#use-environment-variables-in-the-configuration

{{% docs/reference %}}
[local.file_match]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/local.file_match.md"
[local.file_match]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components/local.file_match.md"
[loki.source.file]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.file.md"
[loki.source.file]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.source.file.md"
[loki.write]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.write.md"
[loki.write]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.write.md"
[Components]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/components.md"
[Components]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/concepts/components.md"
[convert]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/cli/convert.md"
[convert]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/cli/convert.md"
[run]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/cli/run.md"
[run]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/cli/run.md"
[Start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/setup/start-agent.md"
[Start]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/setup/start-agent.md"
[DebuggingUI]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging.md"
[DebuggingUI]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging.md"
[River]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/config-language/_index.md"
[River]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/config-language/_index.md"
[UI]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging#grafana-agent-flow-ui"
[UI]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging#grafana-agent-flow-ui"
{{% /docs/reference %}}
