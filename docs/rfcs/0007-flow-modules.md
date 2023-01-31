# Flow Modules

* Date: 2023-01-27
* Author: Matt Durham @mattdurham
* PR: [grafana/agent#XXXX](https://github.com/grafana/agent/pull/XXXX)
* Status: Draft

## RFC Goals

* Explain the use cases of modules
* Explain what modules are
* Go over possible syntax for modules
* Go over pros and cons of modules

## Summary

One of the primary goals for the production usage of Agent Flow is parity with the static subsystem. One of the features of the static subsystem is [scraping service](). Scraping service allows a user to run a fleet of agents and have thousands of scrape configurations distributed among running Agents. Through discussions within the Agent Team, we did not want to limit dynamically loading content to only scrape configs but allow most components to be loaded and used.

During this time the Agent team saw a lot of potential in modules outside of scraping service. Packaging up sets of components for specific workflows, publishing common use cases and allowing better usage for internal customers in the Agent as a Service model.

## Goals

* Enable re-use of common patterns
* Allow loading a module from a string
* Allow modules to load other modules
* Sandbox modules except via arguments and exports
* Allow multiple modules with the same interface to be loaded at once

## Non Goals

* Add additional capabilities to load strings
* Any type of versioning
* Any User Interface work

## Example

```river
# module
argument "password" {
    optional = false
    comment = "password for mysql"
}

argument "username" {
    optional = false
    comment = "username for mysql"
}

export "target" {
    comment = "target for the integration"
    value = integrations.mysql.targets
}

integrations.mysql "mysql" {
    username = argument.username.value
    password = argument.password.value
}
```

```river
# parent

local.file "mysql" {
    filename = "/test/mysql.river"
}

module.single "mysql" {
    content = local.file.mysql.content
    arguments = {
        {
            "password" : PASSWORD,
            "username" : USERNAME,
        }
    }
}

prometheus.scrape "scraper" {
    targets = module.single.mysql.target
}

```

## Limitations

* Duplicate modules cannot be nested, this may or may not be enforced by the system
* Singleton components are not supported at this time
* Modules will not prevent competing resources, such as starting a server on the same port
* Component-like objects will not be supported, ie logging level
* Arguments and exports within a module must be unique

## Proposal

Add the ability to load `modules` as subgraphs to the primary `graph`. Modules may call other modules with a reasonable stack size. Modules are represented as a river string that is interpreted with a defined set of arguments and exports.

The initial component will be `module.single` that will load a single module. Internally these modules will be namespaced so they cannot affect children or parent graphs except via arguments and exports.

Modules will have access to any standard function and any other component exempting singletons. Internally each component in the module will have an `id` that is prepended with the parent's `id` for identification purposes outside of the module. Within the module a component can reference another sibling component normally. There are no known limits on the datatype that a module can use as an argument or export.

### Component Options


Given the above example, the `id` of `integrations.mysql "mysql"` would be `module.single.mysql.integrations.mysql.mysql`. The `data-agent` field would also be prefixed. There are some inherent issues, deeply nested metrics are likely to run into prometheus label value limits. On windows platforms there could be issues with the `data-agent` length. These are issues that currently exist in Agent Flow but are more easily hit using deeply nested modules.


### Failure Modes

#### When a Module Fails Itself and Children Stops

If an error occurs while re-evaluating a module then the module marks itself as unhealthy and unloads the original module.

*Pros*

* Simple to implement
* Easy to understand

*Cons*

* One failure mode can cascade

#### Modules Keep Last Good Value

If an error occurs while re-evaluating a module then the module marks itself as unhealthy and attempts to keep the original module. This may have an issue with cascading failures, if a module depends on a module then the system may enter an inconsistent state while applying and then rolling back the change.

For example, `Module A` has two sub-modules `Module B` and `Module C`. During reevaluation `Module B` reloads appropriately but `Module C` fails. `Module A` unloads both modules and then reloads the last good string. In the case that the last good string also fails then `Module A` is unhealthy and non-functional and `Module A's` submodules do not exist.

*Pros*

* Allows more resilient usage

*Cons*

* Can create undefined behavior
* Complex to unload and reload

## Allowing multiple modules to be loaded at once

Note: This feels the most experimental of the topics listed.

### module.multiple

`module.multiple` will load a module if it matches the required_arguments and required_exports. The input for `source` is a `map(string)`. Note this has a problem of lack of type checking.

Exports are a `map` with the key being the export name and the value of a concatenated array of individual module exports. Exports are accessed via `module.multiple.LABEL.exports.NAME`, `exports` is a `map(array)`.

#### Example

When using the modules for a scraping service there should be the capability to load all modules at once if they have the same arguments and exports.

```river

# Note this doesnt exist and should only be used for representative purposes.
local.files "loadfolder" {
    folder_path = "/configs" # Assume this outputs a map(string)
    filter = "*.river"
}

module.multiple "load" {
    source = local.files.loadfolder.contents
    required_arguments = ["input"]
    required_exports = ["targets"]
}

prometheus.scrape "module" {
    # The module.multiple coalesces multiple exports into an array. 
    targets = module.multiple.load.exports.targets
}

```