---
title: Migrating from Prometheus
weight: 320
---

# Migrating from Prometheus

Migration to Grafana Agent flow mode from [Prometheus][] can be done using
built in Grafana Agent tooling.

This topic describes how to:

* Convert a Prometheus configuration to a flow configuration
* Run a Prometheus configuration natively using Grafana Agent flow mode

[Prometheus]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/

## Components used in this topic

* [prometheus.scrape][]
* [prometheus.remote_write][]

[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md" >}}
[prometheus.remote_write]: {{< relref "../reference/components/prometheus.remote_write.md" >}}

## Before you begin

* Have an existing Prometheus configuration.
* Have a set of Prometheus applications ready to push telemetry data to
  Grafana Agent.
* Be familiar with the concept of [Components][] in Grafana Agent flow mode.

[Components]: {{< relref "../concepts/components.md" >}}
[convert]: {{< relref "../reference/cli/convert.md" >}}
[run]: {{< relref "../reference/cli/run.md" >}}
[Start the agent]: {{< relref "../setup/start-agent.md" >}}
[Flow Debugging]: {{< relref "../monitoring/debugging.md" >}}
[debugging]: #debugging

## Convert a Prometheus configuration

In order to fully migrate your configuration from [Prometheus] to Grafana Agent
Flow, your Prometheus configuration must be converted into a Grafana Agent flow
mode configuration. This will enable you to modify it going forward as a flow
configuration and take full advantage of the many additional features available
in Grafana Agent flow mode.

> In this task, we will use the [convert][] CLI command to output flow
> configuration from a Prometheus configuration.

1. Execute the following:

    ```bash
    grafana-agent convert --format=prometheus --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```
  
    1. Replace `INPUT_CONFIG_PATH` with the full path to the Prometheus configuration.
    2. Replace `OUTPUT_CONFIG_PATH` with the full path to output the flow configuration.

2. [Start the agent][] in flow mode using the new flow configuration from `OUTPUT_CONFIG_PATH`:

The following example demonstrates converting a Prometheus configuration:

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

Execute:

```bash
grafana-agent convert --format=prometheus --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
```

The contents of the new flow configuration:

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

### Debugging

1. If Prometheus configuration is provided that cannot be converted,
   diagnostic information is printed to `stderr`. You can bypass
   any non-critical issues and output the flow configuration using best
   effort conversion by including the `--bypass-errors` flag.
   
    > Be aware that the behavior may not match when bypassing errors.

    ```bash
    grafana-agent convert --format=prometheus --bypass-errors --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

2. You can also output a diagnostic report by including the `--report` flag.

    ```bash
    grafana-agent convert --format=prometheus --report=OUTPUT_REPORT_PATH --output=OUTPUT_CONFIG_PATH INPUT_CONFIG_PATH
    ```

    1. Replace `OUTPUT_REPORT_PATH` with the full path to output the report to.

    Using the example Prometheus configuration from above outputs a diagnostic
    report:

    ```
    (Info) Converted scrape_configs job_name "prometheus" into...
      A prometheus.scrape.prometheus component
    (Info) Converted 1 remote_write[s] "grafana-cloud" into...
      A prometheus.remote_write.default component
    ```

## Run a Prometheus configuration

If you’re not ready to switch over to flow configuration, you can also run
the Prometheus configuration without having to save it as a flow config.
This allows you to try flow mode without having to modify your existing
Prometheus configuration infrastructure.

> In this task, we will use the [run][] CLI command to run the Agent in flow
> mode using a Prometheus configuration.

1. [Start the agent][] in flow mode and include the command line flag
   `--config.format=prometheus`. Your configuration file should be a valid
   Prometheus configuration rather than a flow mode configuration


### Debugging

1. The convert CLI command [debugging][] instructions can be followed to
   generate a diagnostic report.

2. See the Grafana Agent [Flow Debugging][] for debugging a running Grafana
   Agent in flow mode.

3. If the Prometheus configuration provided cannot be converted,
   diagnostic information is printed to `stderr`. You can bypass
   any non-critical issues and start the Agent by including the
   `--config.bypass-conversion-errors` flag in addition to
   `--config.format=prometheus`.

    > Be aware that the behavior may not match when bypassing errors
    > and doing so should be avoided in Production systems.