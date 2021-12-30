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
# Configures the server of the Agent used to enable self-scraping.
[server: <server_config>]

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

NOTE SERVER MUST BE DEFINED IN THE DYNAMIC CONFIGURATION YAML

### Metrics

Metrics are defined using the pattern `metrics-*.yml`, only ONE metrics file is supported.

### Metrics Instances

Metrics Instances are defined using the pattern `metrics_instances-*.yml`, these support more than one file.

## Exporters/Integrations 

Exporters/Integrations are defined using the patten `exporters-*.yml`, these support more than one file, and multiple 
exporters can be defined in a file. NOTE that only one of each exporter is currently supported.