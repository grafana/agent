---
aliases:
- /docs/grafana-cloud/agent/flow/reference/cli/convert/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/cli/convert/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/cli/convert/
canonical: https://grafana.com/docs/agent/latest/flow/reference/cli/convert/
description: The `convert` command converts supported configuration formats to River
  format.
labels:
  stage: beta
menuTitle: convert
title: The convert command
description: Learn about the convert command
weight: 100
---

# The convert command

The `convert` command converts a supported configuration format to Grafana Agent Flow River format.

## Usage

Usage:

* `AGENT_MODE=flow grafana-agent convert [FLAG ...] FILE_NAME`
* `grafana-agent-flow convert [FLAG ...] FILE_NAME`

   Replace the following:

   * `FLAG`: One or more flags that define the input and output of the command.
   * `FILE_NAME`: The Grafana Agent configuration file.

If the `FILE_NAME` argument is not provided or if the `FILE_NAME` argument is
equal to `-`, `convert` converts the contents of standard input. Otherwise,
`convert` reads and converts the file from disk specified by the argument.

There are several different flags available for the `convert` command. You can use the `--output` flag to write the contents of the converted config to a specified path. You can use the `--report` flag to generate a diagnostic report. The `--bypass-errors` flag allows you to bypass any [errors] generated during the file conversion.

The command fails if the source config has syntactically incorrect
configuration or cannot be converted to Grafana Agent Flow River format.

The following flags are supported:

* `--output`, `-o`: The filepath and filename where the output is written.

* `--report`, `-r`: The filepath and filename where the report is written.

* `--source-format`, `-f`: Required. The format of the source file. Supported formats: [prometheus], [promtail], [static].

* `--bypass-errors`, `-b`: Enable bypassing errors when converting.

[prometheus]: #prometheus
[promtail]: #promtail
[static]: #static
[errors]: #errors

### Defaults

Flow Defaults are managed as follows:
* If a provided source config value matches a Flow default value, the property is left off the Flow output.
* If a non-provided source config value default matches a Flow default value, the property is left off the Flow output.
* If a non-provided source config value default doesn't match a Flow default value, the Flow default value is included in the Flow output.

### Errors

Errors are defined as non-critical issues identified during the conversion
where an output can still be generated. These can be bypassed using the
`--bypass-errors` flag.

### Prometheus

Using the `--source-format=prometheus` will convert the source config from
[Prometheus v2.45](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/)
to Grafana Agent Flow config.

This includes Prometheus features such as
[scrape_config](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#scrape_config), 
[relabel_config](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#relabel_config),
[metric_relabel_configs](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#metric_relabel_configs),
[remote_write](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#remote_write),
and many supported *_sd_configs. Unsupported features in a source config result
in [errors].

Refer to [Migrate from Prometheus to Grafana Agent Flow]({{< relref "../../getting-started/migrating-from-prometheus/" >}}) for a detailed migration guide.

### Promtail

Using the `--source-format=promtail` will convert the source configuration from
[Promtail v2.8.x](/docs/loki/v2.8.x/clients/promtail/)
to Grafana Agent Flow configuration.

Nearly all [Promtail features](/docs/loki/v2.8.x/clients/promtail/configuration/)
are supported and can be converted to Grafana Agent Flow config.

If you have unsupported features in a source configuration, you will receive [errors] when you convert to a flow configuration. The converter will
also raise warnings for configuration options that may require your attention.

Refer to [Migrate from Promtail to Grafana Agent Flow]({{< relref "../../getting-started/migrating-from-promtail/" >}}) for a detailed migration guide.

### Static

Using the `--source-format=static` will convert the source configuration from
Grafana Agent [Static]({{< relref "../../../static" >}}) mode to Flow mode configuration.

If you have unsupported features in a Static mode source configuration, you will receive [errors][] when you convert to a Flow mode configuration. The converter will
also raise warnings for configuration options that may require your attention.

Refer to [Migrate Grafana Agent from Static mode to Flow mode]({{< relref "../../getting-started/migrating-from-static/" >}}) for a detailed migration guide.