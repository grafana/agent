---
aliases:
- ../../concepts/custom-components/
- /docs/grafana-cloud/agent/flow/concepts/custom-components/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/concepts/custom-components/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/concepts/custom-components/
- /docs/grafana-cloud/send-data/agent/flow/concepts/custom-components/
canonical: https://grafana.com/docs/agent/latest/flow/concepts/custom-components/
description: Learn about custom components
title: Custom components
weight: 300
---

# Custom components

_Custom components_ are a way to create new components from a pipeline of built-in and other custom components.

A custom component is composed of:

* _Arguments_: settings that configure the custom component.
* _Exports_: values that a custom component exposes to its consumers.
* _Components_: Built-in and custom components that are run as part of the custom component.

## Creating custom components

A new custom component is created using the `declare` configuration block, where the label of the block determines the name of the custom component.

Inside of a `declare` block, the following custom configuration blocks can be used:

* [argument][]: Create a new named argument, whose current value can be referenced using the expression `argument.NAME.value`. Argument values are determined by the user of a custom component.
* [export][]: Expose a new named value to custom component users.

Custom components are useful for reusing a common pipeline multiple times. To learn how to share custom components across multiple files, refer to [Modules][].

[argument]: {{< relref "../reference/config-blocks/argument.md" >}}
[export]: {{< relref "../reference/config-blocks/export.md" >}}
[Modules]: {{< relref "./modules.md" >}}

## Example

This example creates a new custom component called `add` which exports the sum
of two arguments:

```river
declare "add" {
    argument "a" { }
    argument "b" { }

    export "sum" {
        value = argument.a.value + argument.b.value
    }
}

add "example" {
    a = 15
    b = 17
}

// add.example.sum == 32
```
