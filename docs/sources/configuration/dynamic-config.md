---
title: dynamic_config
weight: 500
---

# Dynamic Configuration - Experimental

**This is experimental and subject to change at anytime, feedback is much appreciated. This is a feature that MAY NOT make it production.**

Dynamic Configuration is the combination of two things:

* Loading from multiple files
* Using templates and datasources

Both of these make heavy use of the excellent [gomplate](https://github.com/hairyhenderson/gomplate). The goal is
that as the configuration grows that it can be split it up into smaller segments to allow better readability and handling.
The configurations cannot be patched in any order and instead are allowed at several levels.

The second goal is to allow the use of templating, functions for [gomplate doc](https://docs.gomplate.ca/) go into detail
on what functions are available.

## Configuration

Dynamic configuration files can be used by passing `-config.file.type=dynamic
-enable-features=dynamic-config,integrations-next`. When these flags are
passed, the file referred to `-config.file` will be loaded as a dynamic
configuration file.

Dynamic configuration files are YAML which conform the following schema:

```yaml
# Sources to pull template values
datasources:
  [- <sources_config>]

# Locations to use searching for templates, the system does NOT look into subdirectories. Follows gomplate schema
# from [gomplate datasources](https://docs.gomplate.ca/datasources/). File and S3/GCP templates are currently supported
template_paths:
  [ - string ]

# Filters allow you to override the default naming convention

agent_filter:            string # defaults to agent-*.yml
server_filter:           string # defaults to server-*.yml
metrics_filter:          string # defaults to metrics-*.yml
metrics_instance_filter: string # defaults to metrics_instances-*.yml
integrations_filter:     string # defaults to integrations-*.yml
logs_filter:             string # defaults to logs-*.yml
traces_filter:           string # defaults to traces-*.yml
```

### sources_config
```yaml
# Name of the source to use when templating
name: string

# Path to datasource using schema from [gomplate datasources](https://docs.gomplate.ca/datasources/)
url: string

```

## Templates

Note when adding a template you MUST NOT add the type as the top level yaml field. For instance if using traces:

Incorrect

```yaml
traces:
  configs:
  - name: default
    automatic_logging:
      backend: loki
      loki_name: default
      spans: true
```

Correct

```yaml
configs:
- name: default
  automatic_logging:
    backend: loki
    loki_name: default
    spans: true
```

Configurations are loaded in the order as they are listed below.

### Agent


Agent template is the standard agent configuration file in its entirety. The default filter is `agent-*.yml`. Only
one file is supported. This is processed first then any subsequent configurations found REPLACE the values here, it is
not additive.

[Reference]({{< relref "./" >}})

### Server

The default filter is `server-*.yml`, only ONE server file is supported.

[Reference]({{< relref "./server-config.md" >}})


### Metrics

The default filter is `metrics-*.yml`, only ONE metrics file is supported.

[Reference]({{< relref "./metrics-config.md" >}})

### Metric Instances

The default filter is `metrics_instances-*.yml`. Any metric instances are appended to the instances defined in Metrics above. Any number of metric instance files are supporter.

[Reference]({{< relref "./metrics-config.md#metrics_instance_config" >}}) in the metrics instance


### Integrations

The default filter is `integrations-*.yml`, these support more than one file, and multiple integrations can be defined in a file. Do not assume any order of loading for integrations. For any integration that is a singleton, loading multiple of those will result in an error.

[Reference]({{< relref "./integrations/" >}})

### Traces

The default filter is `traces-*.yml`. This supports ONE file.

[Reference]({{< relref "./traces-config/" >}})

### Logs

The default filter is `logs-*.yml`. This supports ONE file.

[Reference]({{< relref "./logs-config/" >}})
