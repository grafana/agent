---
aliases:
- /docs/agent/latest/concepts/configuration-language
title: Configuration language
weight: 300
---

# Configuration language

The Grafana Agent Flow _configuration language_ refers to the language used in
configuration files which define and configure components to run.

The configuration language is called River, a Terraform/HCL-inspired language:

```river
metrics.scrape "default" {
  targets = [{
    "__address__" = "demo.robustperception.io:9090",
  }]
  forward_to = [metrics.remote_write.default.receiver]
}

metrics.remote_write "default" {
  remote_write {
    url = "http://localhost:9009/api/prom/push"
  }
}
```

River was designed with two requirements in mind:

* _Fast_: The configuration language must be fast so the component controller
  can evaluate changes as quickly as possible.
* _Simple_: The configuration language must be easy to read and write to
  minimize the learning curve.
* _Debuggable_: The configuration language must give detailed information when
  there's a mistake in the config file.

## Attributes

_Attributes_ are used to configure individual settings. They always take the
form of `<ATTRIBUTE_NAME> = <ATTRIBUTE_VALUE>`.

```river
log_level = "debug"
```

This sets the `log_level` attribute to `"debug"`.

## Expressions

Expressions are used to compute the value of an attribute. The simplest
expressions are constant values like `"debug"`, `32`, or `[1, 2, 3, 4]`. River
supports more complex expressions, such as:

* Referencing the exports of components: `local.file.password_file.content`
* Mathematical operations: `1 + 2`, `3 * 4`, `(5 * 6) + (7 + 8)`
* Equality checks: `local.file.file_a.content == local.file.file_b.content`
* Calling functions from River's standard library: `env("HOME")` (retrieve the
  value of the `HOME` environment variable)

Expressions may be used for any attribute inside a component definition.

## Blocks

_Blocks_ are used to configure components and groups of attributes. Each block
can contain any number of attributes or nested blocks.

```river
metrics.remote_write "default" {
  remote_write {
    url = "http://localhost:9009/api/prom/push"
  }
}
```

This file has two blocks:

* `metrics.remote_write "default"`: A labeled block which instantiates a
  `metrics.remote_write` component. The label is the string `"default"`.

* `remote_write`: An unlabeled block inside the component which configures an
  endpoint to send metrics to. This block sets the `url` attribute to specify
  what the endpoint is.

## More information

River is documented in detail in [Configuration language][config-docs] section
of the Grafana Agent Flow docs.

[config-docs]: {{< relref "../config-language/" >}}
