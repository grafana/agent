---
aliases:
- /docs/grafana-cloud/agent/flow/getting-started/migrating-from-static/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/migrating-from-static/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/getting-started/migrating-from-static/
- /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-static/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/migrating-from-static/
description: Learn how to migrate your configuration from Grafana Agent Static to Grafana Agent Flow
menuTitle: Migrate from Static to Flow
title: Migrate Grafana Agent Static to Grafana Agent Flow
weight: 340
refs:
  logs:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/logs-config/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/logs-config/
  static:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/
  prometheus-limitations:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/getting-started/migrating-from-prometheus/#limitations
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-prometheus/#limitations
  agent-management:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/agent-management/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/configuration/agent-management/
  metrics:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/metrics-config/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/static/configuration/metrics-config/
  loki.process:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.process/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.process/
  integrations-next:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/static/configuration/integrations/integrations-next//
  prometheus.scrape:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.scrape/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.scrape/
  loki.write:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.write/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.write/
  local.file_match:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/local.file_match/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/local.file_match/
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
  prometheus.remote_write:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.remote_write/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.remote_write/
  promtail-limitations:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/getting-started/migrating-from-promtail/#limitations
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-promtail/#limitations
  river:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/config-language/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/config-language/
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
  ui:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/#grafana-agent-flow-ui
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/#grafana-agent-flow-ui
  debuggingui:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/
  loki.source.file:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.file/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.source.file/
  start:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/setup/start-agent/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/setup/start-agent/
  convert:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/cli/convert/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/cli/convert/
---

# Migrate from {{% param "PRODUCT_ROOT_NAME" %}} Static to {{% param "PRODUCT_NAME" %}}

The built-in {{< param "PRODUCT_ROOT_NAME" >}} convert command can migrate your [Static](ref:static) configuration to a {{< param "PRODUCT_NAME" >}} configuration.

This topic describes how to:

* Convert a Grafana Agent Static configuration to a {{< param "PRODUCT_NAME" >}} configuration.
* Run a Grafana Agent Static configuration natively using {{< param "PRODUCT_NAME" >}}.

## Components used in this topic

* [prometheus.scrape](ref:prometheus.scrape)
* [prometheus.remote_write](ref:prometheus.remote_write)
* [local.file_match](ref:local.file_match)
* [loki.process](ref:loki.process)
* [loki.source.file](ref:loki.source.file)
* [loki.write](ref:loki.write)

## Before you begin

* You must have an existing Grafana Agent Static configuration.
* You must be familiar with the [Components](ref:components) concept in {{< param "PRODUCT_NAME" >}}.

## Convert a Grafana Agent Static configuration

To fully migrate Grafana Agent [Static](ref:static) to {{< param "PRODUCT_NAME" >}}, you must convert
your Static configuration into a {{< param "PRODUCT_NAME" >}} configuration.
This conversion will enable you to take full advantage of the many additional
features available in {{< param "PRODUCT_NAME" >}}.

> In this task, we will use the [convert](ref:convert) CLI command to output a {{< param "PRODUCT_NAME" >}}
> configuration from a Static configuration.

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
    * `OUTPUT_CONFIG_PATH`: The full path to output the {{< param "PRODUCT_NAME" >}} configuration.

1. [Start](ref:start) {{< param "PRODUCT_NAME" >}} using the new {{< param "PRODUCT_NAME" >}} configuration
   from `OUTPUT_CONFIG_PATH`:

### Debugging

1. If the convert command cannot convert a[Static](ref:static) configuration, diagnostic
   information is sent to `stderr`. You can use the `--bypass-errors` flag to
   bypass any non-critical issues and output the {{< param "PRODUCT_NAME" >}} configuration
   using a best-effort conversion.

   {{% admonition type="caution" %}}
   If you bypass the errors, the behavior of the converted configuration may not match the original[Static](ref:static) configuration. Make sure you fully test the converted configuration before using it in a production environment.
   {{% /admonition %}}

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

   Using the [example](#example) Grafana Agent Static configuration below, the diagnostic
   report provides the following information:

    ```plaintext
    (Warning) Please review your agent command line flags and ensure they are set in your Flow mode config file where necessary.
    ```

## Run a Static mode configuration

If you’re not ready to completely switch to a {{< param "PRODUCT_NAME" >}} configuration, you can run
{{< param "PRODUCT_ROOT_NAME" >}} using your existing Grafana Agent Static configuration.
The `--config.format=static` flag tells {{< param "PRODUCT_ROOT_NAME" >}} to convert your[Static](ref:static)
configuration to {{< param "PRODUCT_NAME" >}} and load it directly without saving the new
configuration. This allows you to try {{< param "PRODUCT_NAME" >}} without modifying your existing
Grafana Agent Static configuration infrastructure.

> In this task, we will use the [run](ref:run) CLI command to run {{< param "PRODUCT_NAME" >}} using a Static configuration.

[Start](ref:start) {{< param "PRODUCT_NAME" >}} and include the command line flag
`--config.format=static`. Your configuration file must be a valid[Static](ref:static)
configuration file.

### Debugging

1. You can follow the convert CLI command [debugging][] instructions to generate
   a diagnostic report.

1. Refer to the {{< param "PRODUCT_NAME" >}} [DebuggingUI](ref:debuggingui) for more information about
   running {{< param "PRODUCT_NAME" >}}.

1. If your[Static](ref:static) configuration can't be converted and loaded directly into
   {{< param "PRODUCT_NAME" >}}, diagnostic information is sent to `stderr`. You can use the `
   --config.bypass-conversion-errors` flag with `--config.format=static` to bypass any
   non-critical issues and start {{< param "PRODUCT_NAME" >}}.

   {{% admonition type="caution" %}}
   If you bypass the errors, the behavior of the converted configuration may not match the original Grafana Agent Static configuration. Do not use this flag in a production environment.
   {{%/admonition %}}

## Example

This example demonstrates converting a[Static](ref:static) configuration file to a {{< param "PRODUCT_NAME" >}} configuration file.

The following[Static](ref:static) configuration file provides the input for the conversion:

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

The new {{< param "PRODUCT_NAME" >}} configuration file looks like this:

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

Configuration conversion is done on a best-effort basis. {{< param "PRODUCT_ROOT_NAME" >}} will issue
warnings or errors where the conversion cannot be performed.

Once the configuration is converted, we recommend that you review
the {{< param "PRODUCT_NAME" >}} configuration file, and verify that it is correct
before starting to use it in a production environment.

Furthermore, we recommend that you review the following checklist:

* The following configuration options aren't available for conversion to {{< param "PRODUCT_NAME" >}}: [Integrations next](ref:integrations-next), [Traces](ref:traces), and [Agent Management](ref:agent-management).
  Any additional unsupported features are returned as errors during conversion.
* There is no gRPC server to configure for {{< param "PRODUCT_NAME" >}}, as any non-default configuration will show as unsupported during the conversion.
* Check if you are using any extra command line arguments with Static that aren't present in your configuration file. For example, `-server.http.address`.
* Check if you are using any environment variables in your [Static](ref:static) configuration.
  These will be evaluated during conversion and you may want to replace them with the {{< param "PRODUCT_NAME" >}} Standard library [env](ref:env) function after conversion.
* Review additional [Prometheus Limitations](ref:prometheus-limitations) for limitations specific to your [Metrics](ref:metrics) configuration.
* Review additional [Promtail Limitations](ref:promtail-limitations) for limitations specific to your [Logs](ref:logs) configuration.
* The logs produced by {{< param "PRODUCT_NAME" >}} mode will differ from those produced by Static.
* {{< param "PRODUCT_ROOT_NAME" >}} exposes the {{< param "PRODUCT_NAME" >}} [UI](ref:ui).

[debugging]: #debugging

