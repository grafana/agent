# Dynamic Configuration - BETA

**This is BETA and subject to change at anytime, feedback is much appreciated**

Dynamic Configuration is the combination of two things:

* Loading from multiple files
* Using templates and datasources

Both of these make heavy use of the excellent [gomplate](https://github.com/hairyhenderson/gomplate). The idea goal is
that as the configuration grows that it can be split it up into smaller segments to allow better readability and handling.
The configurations cannot be patched in any order and instead are allowed at several levels.

The second goal is to allow the use of templating, functions for [gomplate doc](https://docs.gomplate.ca/) go into detail
on what functions are available.

## Configuration

Location of the dynamic configuration is used via the feature flag `dynamic-config`, then it will use `-config.file` to
load the configuration for dynamic configuration.

```yaml
# Sources to pull template values 
[sources: <sources_config>]

# Locations to use searching for templates, the system does NOT look into subdirectories. Follows gomplate schema
# from [gomplate datasources](https://docs.gomplate.ca/datasources/). File and S3/GCP templates are currently supported
templates: 
[ - string ]

``` 

### sources_config
```yaml
# Name of the source to use when templating
name: string

# Path to datasource using schema from [gomplate datasources](https://docs.gomplate.ca/datasources/) 
url: string

```

## Templates

Note when adding a template you do NOT need to add the type as the top level yaml field. For instance if using traces:

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

Agent template is the standard agent configuration file in its entirety. It is defined by the pattern `agent-*.yml`. Only
one file is supported. This is processed first then any subsequent configurations found REPLACE the values here, it is
not additive.

### Server

Serve is dedfined using the pattern `server-*.yml`, only ONE server file is supported.

### Metrics

Metrics are defined using the pattern `metrics-*.yml`, only ONE metrics file is supported.

### Metric Instances

Metric Instances are defined using the pattern `metrics_instances-*.yml`.

### Integrations

Integrations are defined using the pattern `integrations-*.yml`, these support more than one file, and multiple
integrations can be defined in a file. Do not assume any order of loading for integrations.

### Traces

Traces are defined using the pattern `traces-*.yml`. This supports one file.

### Logs

Logs are defined using the pattern `logs-*.yml`. This supports on file.