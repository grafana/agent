---
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/migrating-from-promtail/
description: Learn how to migrate your configuration from Promtail to Grafana Agent
  Flow Mode
menuTitle: Migrate from Promtail
title: Migrate from Promtail to Grafana Agent Flow Mode
weight: 320
---

# Migrate from Promtail to Grafana Agent

The built-in Grafana Agent convert command can migrate your [Promtail][]
configuration to a Grafana Agent flow configuration.

This topic describes how to:

* Convert a Promtail configuration to a Flow Mode configuration.
* Run a Promtail configuration natively using Grafana Agent Flow Mode.

[Promtail]: https://grafana.com/docs/loki/latest/clients/promtail/

## Components used in this topic

* [local.file_match][]
* [loki.source.file][]
* [loki.write][]

[local.file_match]: {{< relref "../reference/components/local.file_match.md" >}}
[loki.source.file]: {{< relref "../reference/components/loki.source.file.md" >}}
[loki.write]: {{< relref "../reference/components/loki.write.md" >}}

## Before you begin

* You must have an existing Promtail configuration.
* You must be familiar with the concept of [Components][] in Grafana Agent Flow
  Mode.

[Components]: {{< relref "../concepts/components.md" >}}
[convert]: {{< relref "../reference/cli/convert.md" >}}
[run]: {{< relref "../reference/cli/run.md" >}}
[Start the agent]: {{< relref "../setup/start-agent.md" >}}
[Flow Debugging]: {{< relref "../monitoring/debugging.md" >}}
[debugging]: #debugging

## Convert a Promtail configuration

To fully migrate from [Promtail] to Grafana Agent in Flow Mode, you must convert
your Promtail configuration into a Grafana Agent Flow Mode configuration. This
conversion will enable you to take full advantage of the many additional
features available in Grafana Agent Flow Mode.

> In this task, we will use the [convert][] CLI command to output a flow
> configuration from a Promtail configuration.

1. Open a terminal window and run the following command:

    ```bash
    AGENT_MODE=flow; grafana-agent convert --source-format=promtail --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

   Replace the following:
    * `INPUT_CONFIG_PATH`: The full path to the Promtail configuration.
    * `OUTPUT_CONFIG_PATH`: The full path to output the flow configuration.

1. [Start the agent][] in Flow Mode using the new flow configuration
   from `OUTPUT_CONFIG_PATH`:

### Debugging

1. If the convert command cannot convert a Promtail configuration, diagnostic
   information is sent to `stderr`. You can bypass any non-critical issues and
   output the flow configuration using a best-effort conversion by including
   the `--bypass-errors` flag.

   {{% admonition type="caution" %}}If you bypass the errors, the behavior of the converted configuration may not match the original Promtail configuration. Make sure you fully test the converted configuration before using it in a production environment.{{% /admonition %}}

    ```bash
    AGENT_MODE=flow; grafana-agent convert --source-format=promtail --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

1. You can also output a diagnostic report by including the `--report` flag.

    ```bash
    AGENT_MODE=flow; grafana-agent convert --source-format=promtail --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

    * Replace `OUTPUT_REPORT_PATH` with the output path for the report.

   Using the [example](#example) Promtail configuration below, the diagnostic
   report provides the following information:

    ```plaintext
    (Warning) If you have a tracing set up for Promtail, it cannot be migrated to Flow Mode automatically - please refer to documentation on how to configure tracing in Flow Mode
    (Warning) The Agent Flow Mode's metrics are different from the metrics emitted by Promtail. If you rely on Promtail's metrics you will need to change your configuration (e.g. alerts and dashboards).
    ```

## Run a Promtail configuration

If you’re not ready to completely switch to a flow configuration, you can run
Grafana Agent using your existing Promtail configuration.
The `--config.format=promtail` flag tells Grafana Agent to convert your Promtail
configuration to Flow Mode and load it directly without saving the new
configuration. This allows you to try Flow Mode without modifying your existing
Promtail configuration infrastructure.

> In this task, we will use the [run][] CLI command to run Grafana Agent in Flow
> Mode using a Promtail configuration.

[Start the agent][] in Flow Mode and include the command line flag
`--config.format=promtail`. Your configuration file must be a valid Promtail
configuration file rather than a Flow Mode configuration file.

### Debugging

1. You can follow the convert CLI command [debugging][] instructions to generate
   a diagnostic report.

1. Refer to the Grafana Agent [Flow Debugging][] for more information about a
   running Grafana Agent in Flow Mode.

1. If your Promtail configuration cannot be converted and loaded directly into
   Grafana Agent, diagnostic information is sent to `stderr`. You can bypass any
   non-critical issues and start the Agent by including the
   `--config.bypass-conversion-errors` flag in addition to
   `--config.format=promtail`.

   {{% admonition type="caution" %}}If you bypass the errors, the behavior of
   the converted configuration may not match the original Promtail
   configuration. Do not use this flag in a production environment.{{%
   /admonition %}}

## Example

This example demonstrates converting a Promtail configuration file to a Grafana
Agent Flow Mode configuration file.

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

The convert command takes the YAML file as input and outputs a River file.

```bash
AGENT_MODE=flow; grafana-agent convert --source-format=promtail --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

The new Flow Mode configuration file looks like this:

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

Config conversion is done on the best-effort basis; the Agent will issue
warnings or errors where the conversion cannot be performed.

Once the configuration is converted, it is recommended for users to understand
and review the output Flow Mode configuration and verify that it is correct
before starting to use it in a production environment.

Furthermore, it is recommended to go through the following checklist:

* Check if you are using any extra command line arguments with Promtail which
  are not present in your config file, e.g. `-max-line-size`
* Check if you are setting any environment variables -
  whether [expanded in the config file][] itself or consumed directly by
  Promtail such as `JAEGER_AGENT_HOST`.
* In Flow Mode, the positions file will be saved at a different location -
  see [loki.source.file][] for more details. Check if you have any existing
  setup (e.g. Kubernetes Persistent Volume) that needs changing to use the new
  positions file path.
* Metrics exposed by the Flow Mode will usually match those of Promtail, but
  will use a different name. Make sure that you use the new metric names, e.g.
  in your alerts and dashboards queries.
* Please note that the logs produced by the Agent will differ from those
  produced by Promtail.
* Please note that the Agent exposes [Grafana Agent Flow UI][], which differs
  from Promtail's Web UI.

[expanded in the config file]: https://grafana.com/docs/loki/latest/clients/promtail/configuration/#use-environment-variables-in-the-configuration

[Grafana Agent Flow UI]: https://grafana.com/docs/agent/latest/flow/monitoring/debugging/#grafana-agent-flow-ui
