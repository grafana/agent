---
aliases:
- ../../concepts/configuration-language/
- /docs/grafana-cloud/agent/flow/concepts/configuration_language/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/concepts/configuration_language/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/concepts/configuration_language/
- /docs/grafana-cloud/send-data/agent/flow/concepts/configuration_language/
canonical: https://grafana.com/docs/agent/latest/flow/concepts/configuration_language/
description: Learn about configuration language concepts
title: Configuration language concepts
weight: 400
---

# Configuration language concepts

The {{< param "PRODUCT_NAME" >}} _configuration language_, River, refers to the language used in configuration files that define and configure components to run.

```river
prometheus.scrape "default" {
  targets = [{
    "__address__" = "demo.robustperception.io:9090",
  }]
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9009/api/prom/push"
  }
}
```

River is designed with the following requirements in mind:

* _Fast_: The configuration language must be fast so the component controller can quickly evaluate changes.
* _Simple_: The configuration language must be easy to read and write to minimize the learning curve.
* _Debuggable_: The configuration language must give detailed information when there's a mistake in the configuration file.

## Attributes

You use _Attributes_ to configure individual settings.
Attributes always take the form of `ATTRIBUTE_NAME = ATTRIBUTE_VALUE`.

The following example shows how to set the `log_level` attribute to `"debug"`.

```river
log_level = "debug"
```

## Expressions

You use expressions to compute the value of an attribute.
The simplest expressions are constant values like `"debug"`, `32`, or `[1, 2, 3, 4]`.
River supports complex expressions, for example:

* Referencing the exports of components: `local.file.password_file.content`
* Mathematical operations: `1 + 2`, `3 * 4`, `(5 * 6) + (7 + 8)`
* Equality checks: `local.file.file_a.content == local.file.file_b.content`
* Calling functions from River's standard library: `env("HOME")` retrieves the value of the `HOME` environment variable.

You can use expressions for any attribute inside a component definition.

### Referencing component exports

The most common expression is to reference the exports of a component, for example, `local.file.password_file.content`.
You form a reference to a component's exports by merging the component's name (for example, `local.file`),
label (for example, `password_file`), and export name (for example, `content`), delimited by a period.

## Blocks

You use _Blocks_ to configure components and groups of attributes.
Each block can contain any number of attributes or nested blocks.

```river
prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9009/api/prom/push"
  }
}
```

The preceding example has two blocks:

* `prometheus.remote_write "default"`: A labeled block which instantiates a `prometheus.remote_write` component.
  The label is the string `"default"`.
* `endpoint`: An unlabeled block inside the component that configures an endpoint to send metrics to.
  This block sets the `url` attribute to specify the endpoint.

## More information

Refer to [Configuration language][config-docs] for more information about River.

{{% docs/reference %}}
[config-docs]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/config-language"
[config-docs]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/config-language"
{{% /docs/reference %}}
