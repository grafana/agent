---
aliases:
- /docs/grafana-cloud/agent/flow/reference/config-blocks/argument/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/config-blocks/argument/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/config-blocks/argument/
- /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/argument/
canonical: https://grafana.com/docs/agent/latest/flow/reference/config-blocks/argument/
description: Learn about the argument configuration block
menuTitle: argument
title: argument block
---

# argument block

`argument` is an optional configuration block used to specify parameterized
input to a [Module][Modules]. `argument` blocks must be given a label which
determines the name of the argument.

The `argument` block may not be specified in the main configuration file given
to {{< param "PRODUCT_NAME" >}}.

[Modules]: {{< relref "../../concepts/modules.md" >}}

## Example

```river
argument "ARGUMENT_NAME" {}
```

## Arguments

> **NOTE**: For clarity, "argument" in this section refers to arguments which
> can be given to the argument block. "Module argument" refers to the argument
> being defined for a module, determined by the label of the argument block.

The following arguments are supported:

Name       | Type     | Description                          | Default | Required
-----------|----------|--------------------------------------|---------|---------
`comment`  | `string` | Description for the argument.        | `false` | no
`default`  | `any`    | Default value for the argument.      | `null`  | no
`optional` | `bool`   | Whether the argument may be omitted. | `false` | no

By default, all module arguments are required. The `optional` argument can be
used to mark the module argument as optional. When `optional` is `true`, the
initial value for the module argument is specified by `default`.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`value` | `any` | The current value of the argument.

The module loader is responsible for determining the values for arguments.
Components in a module may use `argument.ARGUMENT_NAME.value` to retrieve the
value provided by the module loader.

## Example

This example creates a module where {{< param "PRODUCT_NAME" >}} metrics are collected. Collected
metrics are then forwarded to the argument specified by the loader:

```river
argument "metrics_output" {
  optional = false
  comment  = "Where to send collected metrics."
}

prometheus.scrape "selfmonitor" {
  targets = [{
    __address__ = "127.0.0.1:12345",
  }]

  forward_to = [argument.metrics_output.value]
}
```
