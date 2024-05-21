---
aliases:
- /docs/grafana-cloud/agent/flow/getting-started/migrating-from-static/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/migrating-from-static/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/getting-started/migrating-from-static/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/migrating-from-static/
description: Learn how to migrate your configuration from Grafana Agent Static mode to Flow mode
menuTitle: Migrate from Static mode to Flow mode
title: Migrate Grafana Agent from Static mode to Flow mode
weight: 340
refs:
  components:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/concepts/components/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/concepts/components/
  env:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/stdlib/env/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/stdlib/env/
  river:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/config-language/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/config-language/
  promtail-limitations:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/getting-started/migrating-from-promtail/#limitations
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-promtail/#limitations
  prometheus.remote_write:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.remote_write/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.remote_write/
  traces:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/traces-config/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/configuration/traces-config/
  run:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/cli/run/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/cli/run/
  convert:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/cli/convert/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/cli/convert/
  loki.source.file:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.file/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.source.file/
  loki.process:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.process/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.process/
  metrics:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/metrics-config/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/configuration/metrics-config/
  integrations-next:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/integrations/integrations-next//
  start-the-agent:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/setup/start-agent/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/setup/start-agent/
  prometheus.scrape:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.scrape/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.scrape/
  local.file_match:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/local.file_match/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/local.file_match/
  loki.write:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.write/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.write/
  grafana-agent-flow-ui:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/#grafana-agent-flow-ui
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/#grafana-agent-flow-ui
  prometheus-limitations:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/getting-started/migrating-from-prometheus/#limitations
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-prometheus/#limitations
  static:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/
  agent-management:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/agent-management/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/configuration/agent-management/
  logs:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/logs-config/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/logs-config/
  flow-debugging:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/
---

# Migrate Grafana Agent from Static mode to Flow mode

The built-in Grafana Agent convert command can migrate your [Static](ref:static) mode
configuration to a Flow mode configuration.

This topic describes how to:

* Convert a Grafana Agent Static mode configuration to a Flow mode configuration.
* Run a Grafana Agent Static mode configuration natively using Grafana Agent Flow mode.

## Components used in this topic

* [prometheus.scrape](ref:prometheus.scrape)
* [prometheus.remote_write](ref:prometheus.remote_write)
* [local.file_match](ref:local.file_match)
* [loki.process](ref:loki.process)
* [loki.source.file](ref:loki.source.file)
* [loki.write](ref:loki.write)

## Before you begin

* You must have an existing Grafana Agent Static mode configuration.
* You must be familiar with the [Components](ref:components) concept in Grafana Agent Flow mode.

## Convert a Static mode configuration

To fully migrate Grafana Agent from [Static](ref:static) mode to Flow mode, you must convert
your Static mode configuration into a Flow mode configuration.
This conversion will enable you to take full advantage of the many additional
features available in Grafana Agent Flow mode.

> In this task, we will use the [convert](ref:convert) CLI command to output a Flow mode
> configuration from a Static mode configuration.

1. Open a terminal window and run the following command:

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=static --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=static --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   {{< /code >}}

   Replace the following:
    * `INPUT_CONFIG_PATH`: The full path to the [Static](ref:static) configuration.
    * `OUTPUT_CONFIG_PATH`: The full path to output the flow configuration.

1. [Start the Agent][] in Flow mode using the new Flow mode configuration
   from `OUTPUT_CONFIG_PATH`:

### Debugging

1. If the convert command cannot convert a[Static](ref:static) mode configuration, diagnostic
   information is sent to `stderr`. You can use the `--bypass-errors` flag to 
   bypass any non-critical issues and output the Flow mode configuration 
   using a best-effort conversion.

   {{% admonition type="caution" %}}If you bypass the errors, the behavior of the converted configuration may not match the original[Static](ref:static) mode configuration. Make sure you fully test the converted configuration before using it in a production environment.{{% /admonition %}}

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=static --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=static --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   {{< /code >}}

1. You can use the `--report` flag to output a diagnostic report.

   {{< code >}}

   ```static-binary
   AGENT_MODE=flow grafana-agent convert --source-format=static --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
   ```

   ```flow-binary
   grafana-agent-flow convert --source-format=static --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

   {{< /code >}}

    * Replace `OUTPUT_REPORT_PATH` with the output path for the report.

   Using the [example](#example) Grafana Agent Static Mode configuration below, the diagnostic
   report provides the following information:

    ```plaintext
    (Warning) Please review your agent command line flags and ensure they are set in your Flow mode config file where necessary.
    ```

## Run a Static mode configuration

If you’re not ready to completely switch to a Flow mode configuration, you can run
Grafana Agent using your existing Grafana Agent Static mode configuration.
The `--config.format=static` flag tells Grafana Agent to convert your[Static](ref:static) mode
configuration to Flow mode and load it directly without saving the new
configuration. This allows you to try Flow mode without modifying your existing
Grafana Agent Static mode configuration infrastructure.

> In this task, we will use the [run](ref:run) CLI command to run Grafana Agent in Flow
> mode using a Static mode configuration.

[Start the Agent][] in Flow mode and include the command line flag
`--config.format=static`. Your configuration file must be a valid[Static](ref:static)
mode configuration file.

### Debugging

1. You can follow the convert CLI command [debugging][] instructions to generate
   a diagnostic report.

1. Refer to the Grafana Agent [Flow Debugging](ref:flow-debugging) for more information about
   running Grafana Agent in Flow mode.

1. If your[Static](ref:static) mode configuration can't be converted and loaded directly into
    Grafana Agent, diagnostic information is sent to `stderr`. You can use the `
    --config.bypass-conversion-errors` flag with `--config.format=static` to bypass any
    non-critical issues and start the Agent.

   {{% admonition type="caution" %}}If you bypass the errors, the behavior of the converted configuration may not match the original Grafana Agent Static mode configuration. Do not use this flag in a production environment.{{%/admonition %}}

## Example

This example demonstrates converting a[Static](ref:static) mode configuration file to a Flow mode configuration file.

The following[Static](ref:static) mode configuration file provides the input for the conversion:

```yaml
server:
  log_level: info

metrics:
  global:
    scrape_interval: 15s
    remote_write:
      - url: https://prometheus-us-central1.grafana.net/api/prom/push
        basic_auth:
          username: USERNAME
          password: PASSWORD
  configs:
    - name: test
      host_filter: false
      scrape_configs:
        - job_name: local-agent
          static_configs:
            - targets: ['127.0.0.1:12345']
              labels:
                cluster: 'localhost'

logs:
  global:
    file_watch_config:
      min_poll_frequency: 1s
      max_poll_frequency: 5s
  positions_directory: /var/lib/agent/data-agent
  configs:
    - name: varlogs
      scrape_configs:
        - job_name: varlogs
          static_configs:
            - targets:
              - localhost
              labels:
                job: varlogs
                host: mylocalhost
                __path__: /var/log/*.log
          pipeline_stages:
            - match:
                selector: '{filename="/var/log/*.log"}'
                stages:
                - drop:
                    expression: '^[^0-9]{4}'
                - regex:
                    expression: '^(?P<timestamp>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \[(?P<level>[[:alpha:]]+)\] (?:\d+)\#(?:\d+): \*(?:\d+) (?P<message>.+)$'
                - pack:
                    labels:
                      - level
      clients:
        - url: https://USER_ID:API_KEY@logs-prod3.grafana.net/loki/api/v1/push
```

The convert command takes the YAML file as input and outputs a [River](ref:river) file.

{{< code >}}

```static-binary
AGENT_MODE=flow grafana-agent convert --source-format=static --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

```flow-binary
grafana-agent-flow convert --source-format=static --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

{{< /code >}}

The new Flow mode configuration file looks like this:

```river
prometheus.scrape "metrics_test_local_agent" {
	targets = [{
		__address__ = "127.0.0.1:12345",
		cluster     = "localhost",
	}]
	forward_to      = [prometheus.remote_write.metrics_test.receiver]
	job_name        = "local-agent"
	scrape_interval = "15s"
}

prometheus.remote_write "metrics_test" {
	endpoint {
		name = "test-3a2a1b"
		url  = "https://prometheus-us-central1.grafana.net/api/prom/push"

		basic_auth {
			username = "USERNAME"
			password = "PASSWORD"
		}

		queue_config { }

		metadata_config { }
	}
}

local.file_match "logs_varlogs_varlogs" {
	path_targets = [{
		__address__ = "localhost",
		__path__    = "/var/log/*.log",
		host        = "mylocalhost",
		job         = "varlogs",
	}]
}

loki.process "logs_varlogs_varlogs" {
	forward_to = [loki.write.logs_varlogs.receiver]

	stage.match {
		selector = "{filename=\"/var/log/*.log\"}"

		stage.drop {
			expression = "^[^0-9]{4}"
		}

		stage.regex {
			expression = "^(?P<timestamp>\\d{4}/\\d{2}/\\d{2} \\d{2}:\\d{2}:\\d{2}) \\[(?P<level>[[:alpha:]]+)\\] (?:\\d+)\\#(?:\\d+): \\*(?:\\d+) (?P<message>.+)$"
		}

		stage.pack {
			labels           = ["level"]
			ingest_timestamp = false
		}
	}
}

loki.source.file "logs_varlogs_varlogs" {
	targets    = local.file_match.logs_varlogs_varlogs.targets
	forward_to = [loki.process.logs_varlogs_varlogs.receiver]

	file_watch {
		min_poll_frequency = "1s"
		max_poll_frequency = "5s"
	}
}

loki.write "logs_varlogs" {
	endpoint {
		url = "https://USER_ID:API_KEY@logs-prod3.grafana.net/loki/api/v1/push"
	}
	external_labels = {}
}
```

## Limitations

Configuration conversion is done on a best-effort basis. The Agent will issue
warnings or errors where the conversion cannot be performed.

Once the configuration is converted, we recommend that you review
the Flow mode configuration file, and verify that it is correct
before starting to use it in a production environment.

Furthermore, we recommend that you review the following checklist:

* The following configuration options are not available for conversion to Flow
  mode: [Integrations next](ref:integrations-next), [Traces](ref:traces), and [Agent Management](ref:agent-management). Any
  additional unsupported features are returned as errors during conversion.
* There is no gRPC server to configure for Flow mode, so any non-default config
  will show as unsupported during the conversion.
* Check if you are using any extra command line arguments with Static mode that
  are not present in your configuration file. For example, `-server.http.address`.
* Check if you are using any environment variables in your[Static](ref:static) mode configuration.
  These will be evaluated during conversion and you may want to replace them
  with the Flow Standard library[env](ref:env) function after conversion.
* Review additional[Prometheus Limitations](ref:prometheus-limitations) for limitations specific to your
 [Metrics](ref:metrics) config.
* Review additional[Promtail Limitations](ref:promtail-limitations) for limitations specific to your
 [Logs](ref:logs) config.
* The logs produced by Grafana Agent Flow mode will differ from those
  produced by Static mode.
* Grafana Agent exposes the [Grafana Agent Flow UI](ref:grafana-agent-flow-ui).

[debugging]: #debugging

