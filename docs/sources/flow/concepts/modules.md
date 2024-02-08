---
aliases:
- ../../concepts/modules/
- /docs/grafana-cloud/agent/flow/concepts/modules/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/concepts/modules/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/concepts/modules/
- /docs/grafana-cloud/send-data/agent/flow/concepts/modules/
canonical: https://grafana.com/docs/agent/latest/flow/concepts/modules/
description: Learn about modules
title: Modules
weight: 300
---

# Modules

You use _Modules_ to create {{< param "PRODUCT_NAME" >}} configurations that you can load as a component.
Modules are a great way to parameterize a configuration to create reusable pipelines.

Modules are {{< param "PRODUCT_NAME" >}} configurations which have:

* _Arguments_: Settings that configure a module.
* _Exports_: Named values that a module exposes to the consumer of the module.
* _Components_: {{< param "PRODUCT_NAME" >}} components to run when the module is running.

You use a [Module loader][] to load Modules into {{< param "PRODUCT_NAME" >}}.

Refer to [argument block][] and [export block][] to learn how to define arguments and exports for a module.

## Module loaders

A _Module loader_ is a {{< param "PRODUCT_NAME" >}} component that retrieves a module and runs the defined components.

Module loader components are responsible for the following functions:

* Retrieving the module source.
* Creating a [Component controller][] for the module.
* Passing arguments to the loaded module.
* Exposing exports from the loaded module.

Module loaders are typically called `module.LOADER_NAME`.

{{< admonition type="note" >}}
Some module loaders may not support running modules with arguments or exports.
{{< /admonition >}}

Refer to [Components][] for more information about the module loader components.

## Module sources

Modules are flexible, and you can retrieve their configuration anywhere, such as:

* The local filesystem.
* An S3 bucket.
* An HTTP endpoint.

Each module loader component supports different ways of retrieving `module.sources`.
The most generic module loader component, `module.string`, can load modules from the export of another {{< param "PRODUCT_NAME" >}} component.

```river
local.file "my_module" {
  filename = "PATH_TO_MODULE"
}

module.string "my_module" {
  content = local.file.my_module.content

  arguments {
    MODULE_ARGUMENT_NAME_1 = MODULE_ARGUMENT_VALUE_1
    MODULE_ARGUMENT_NAME_2 = MODULE_ARGUMENT_VALUE_2
    // ...
  }
}
```

## Example module

This example module manages a pipeline that filters out debug-level and info-level log lines.

```river
// argument.write_to is a required argument that specifies where filtered
// log lines are sent.
//
// The value of the argument is retrieved in this file with
// argument.write_to.value.
argument "write_to" {
  optional = false
}

// loki.process.filter is our component which executes the filtering, passing
// filtered logs to argument.write_to.value.
loki.process "filter" {
  // Drop all debug- and info-level logs.
  stage.match {
    selector = "{job!=\"\"} |~ \"level=(debug|info)\""
    action   = "drop"
  }

  // Send processed logs to our argument.
  forward_to = argument.write_to.value
}

// export.filter_input exports a value to the module consumer.
export "filter_input" {
  // Expose the receiver of loki.process so the module consumer can send
  // logs to our loki.process component.
  value = loki.process.filter.receiver
}
```

You can save the module to a file and then use it as a processing step before writing logs to Loki.

```river
loki.source.file "self" {
  targets = LOG_TARGETS

  // Forward collected logs to the input of our filter.
  forward_to = [module.file.log_filter.exports.filter_input]
}

module.file "log_filter" {
  filename = "/path/to/modules/log_filter.river"

  arguments {
    // Configure the filter to forward filtered logs to loki.write below.
    write_to = [loki.write.default.receiver],
  }
}

loki.write "default" {
  endpoint {
    url = "LOKI_URL"
  }
}
```

[Module loader]: #module-loaders

{{% docs/reference %}}
[argument block]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/argument.md"
[argument block]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/argument.md"
[export block]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/export.md"
[export block]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/export.md"
[Component controller]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/component_controller.md"
[Component controller]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/concepts/component_controller.md"
[Components]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components"
[Components]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components"
{{% /docs/reference %}}
