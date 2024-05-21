---
aliases:
- ../../concepts/modules/
- /docs/grafana-cloud/agent/flow/concepts/modules/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/concepts/modules/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/concepts/modules/
canonical: https://grafana.com/docs/agent/latest/flow/concepts/modules/
title: Modules
description: Learn about modules
weight: 300
refs:
  component-controller:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/concepts/component_controller/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/concepts/component_controller/
  components:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/components/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/components/
  export-block:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/export/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/export/
  argument-block:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/argument/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/argument/
---

# Modules

_Modules_ are a way to create Grafana Agent Flow configurations which can be
loaded as a component. Modules are a great way to parameterize a configuration
to create reusable pipelines.

Modules are Grafana Agent Flow configurations which have:

* Arguments: settings which configure a module.
* Exports: named values which a module exposes to the consumer of the module.
* Components: Grafana Agent Flow Components to run when the module is running.

Modules are loaded into Grafana Agent Flow by using a [Module
loader](#module-loaders).

Refer to the documentation for the [argument block](ref:argument-block) and [export block](ref:export-block) to
learn how to define arguments and exports for a module.

## Module loaders

A _Module loader_ is a Grafana Agent Flow component which retrieves a module
and runs the components defined inside of it.

Module loader components are responsible for:

* Retrieving the module source to run.
* Creating a [Component controller](ref:component-controller) for the module to run in.
* Passing arguments to the loaded module.
* Exposing exports from the loaded module.

Module loaders typically are called `module.LOADER_NAME`. The list of module
loader components can be found in the list of Grafana Agent Flow
[Components](ref:components).

Some module loaders may not support running modules with arguments or exports.
Refer to the documentation for the module loader you are using for more
information.

## Module sources

Modules are designed to be flexible, and can have their configuration retrieved
from anywhere, such as:

* The local filesystem
* An S3 bucket
* An HTTP endpoint

Each module loader component will support different ways of retrieving module
sources. The most generic module loader component, `module.string`, can load
modules from the export of another Flow component:

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

This example module manages a pipeline which filters out debug- and info-level
log lines which are given to it:

```river
// argument.write_to is a required argument which specifies where filtered
// log lines should be sent.
//
// The value of the argument can be retrieved in this file with
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

// export.filter_input exports a value to the consumer of the module.
export "filter_input" {
  // Expose the receiver of loki.process so the module consumer can send
  // logs to our loki.process component.
  value = loki.process.filter.receiver
}
```

The module above can be saved to a file and then used as a processing step
before writing logs to Loki:

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

