---
description: Learn how to migrate your configuration from Prometheus to Grafana Agent flow mode
title: Migrate from Prometheus to Grafana Agent flow mode
menuTitle: Migrate from Prometheus
weight: 320
---

# Migrate from Prometheus to Grafana Agent

The built-in Grafana Agent convert command can migrate your [Prometheus][] configuration to a Grafana Agent flow configuration.

This topic describes how to:

* Convert a Prometheus configuration to a flow configuration.
* Run a Prometheus configuration natively using Grafana Agent flow mode.

[Prometheus]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/

## Components used in this topic

* [prometheus.scrape][]
* [prometheus.remote_write][]

[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md" >}}
[prometheus.remote_write]: {{< relref "../reference/components/prometheus.remote_write.md" >}}

## Before you begin

* You must have an existing Prometheus configuration.
* You must have a set of Prometheus applications ready to push telemetry data to Grafana Agent.
* You must be familiar with the concept of [Components][] in Grafana Agent flow mode.

[Components]: {{< relref "../concepts/components.md" >}}
[convert]: {{< relref "../reference/cli/convert.md" >}}
[run]: {{< relref "../reference/cli/run.md" >}}
[Start the agent]: {{< relref "../setup/start-agent.md" >}}
[Flow Debugging]: {{< relref "../monitoring/debugging.md" >}}
[debugging]: #debugging

## Convert a Prometheus configuration

To fully migrate your configuration from [Prometheus] to Grafana Agent
in flow mode, you must convert your Prometheus configuration into a Grafana Agent flow
mode configuration. This conversion will enable you to take full advantage of the many 
additional features available in Grafana Agent flow mode.

> In this task, we will use the [convert][] CLI command to output a flow
> configuration from a Prometheus configuration.

1. Open a terminal window and run the following command:

    ```bash
    grafana-agent convert --format=prometheus --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```
  
    Replace the following: 
      * `INPUT_CONFIG_PATH`: The full path to the Prometheus configuration.
      * `OUTPUT_CONFIG_PATH`: The full path to output the flow configuration.

1. [Start the agent][] in flow mode using the new flow configuration from `OUTPUT_CONFIG_PATH`:

### Debugging

1. If the convert command cannot convert a Prometheus configuration,
   diagnostic information is sent to `stderr`. You can bypass
   any non-critical issues and output the flow configuration using a 
   best-effort conversion by including the `--bypass-errors` flag.

    {{% admonition type="caution" %}}If you bypass the errors, the behavior of the converted configuration may not match the original Prometheus configuration. Make sure you fully test the converted configuration before using it in a production environment.{{% /admonition %}}

    ```bash
    grafana-agent convert --format=prometheus --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

1. You can also output a diagnostic report by including the `--report` flag.

    ```bash
    grafana-agent convert --format=prometheus --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

    * Replace `OUTPUT_REPORT_PATH` with the output path for the report.

    Using the example Prometheus configuration from above the diagnostic report provides the following information:

    ```plaintext
    (Info) Converted scrape_configs job_name "prometheus" into...
      A prometheus.scrape.prometheus component
    (Info) Converted 1 remote_write[s] "grafana-cloud" into...
      A prometheus.remote_write.default component
    ```

## Run a Prometheus configuration

If you’re not ready to completely switch to a flow configuration, you can run Grafana Agent using your existing Prometheus configuration.
The `--config.format=prometheus` flag tells Grafana Agent to convert your Prometheus configuration to flow mode and load it directly without saving the new configuration.
This allows you to try flow mode without modifying your existing Prometheus configuration infrastructure.

> In this task, we will use the [run][] CLI command to run Grafana Agent in flow
> mode using a Prometheus configuration.

[Start the agent][] in flow mode and include the command line flag
   `--config.format=prometheus`. Your configuration file must be a valid
   Prometheus configuration file rather than a flow mode configuration file.

### Debugging

1. You can follow the convert CLI command [debugging][] instructions to
   generate a diagnostic report.

1. Refer to the Grafana Agent [Flow Debugging][] for more information about a running Grafana
   Agent in flow mode.

1. If your Prometheus configuration cannot be converted and 
    loaded directly into Grafana Agent, diagnostic information 
    is sent to `stderr`. You can bypass any non-critical issues 
    and start the Agent by including the
   `--config.bypass-conversion-errors` flag in addition to
   `--config.format=prometheus`.

    {{% admonition type="caution" %}}If you bypass the errors, the behavior of the converted configuration may not match the original Prometheus configuration. Do not use this flag in a production environment.{{% /admonition %}}

## Example

This example demonstrates converting a Prometheus configuration file to a Grafana Agent flow mode configuration file.

The following Prometheus configuration file provides the input for the conversion:

```yaml
global:
  scrape_timeout:    45s

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:12345"]

remote_write:
  - name: "grafana-cloud"
    url: "https://prometheus-us-central1.grafana.net/api/prom/push"
    basic_auth:
      username: USERNAME
      password: PASSWORD
```

The convert command takes the YAML file as input and outputs a River file.

```bash
grafana-agent convert --format=prometheus --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

The new flow configuration file looks like this:

```river
prometheus.scrape "prometheus" {
  targets = [{
    __address__ = "localhost:12345",
  }]
  forward_to     = [prometheus.remote_write.default.receiver]
  job_name       = "prometheus"
  scrape_timeout = "45s"
}

prometheus.remote_write "default" {
  endpoint {
    name = "grafana-cloud"
    url  = "https://prometheus-us-central1.grafana.net/api/prom/push"

    basic_auth {
      username = "USERNAME"
      password = "PASSWORD"
    }

    queue_config {
      capacity             = 2500
      max_shards           = 200
      max_samples_per_send = 500
    }

    metadata_config {
      max_samples_per_send = 500
    }
  }
}
```
