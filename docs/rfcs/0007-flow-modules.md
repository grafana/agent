# Flow Modules

* Date: 2023-01-27
* Author: Matt Durham @mattdurham
* PR: [grafana/agent#2898](https://github.com/grafana/agent/pull/2898)

[Formatted Link for ease of user](https://github.com/grafana/agent/blob/rfc_modules/docs/rfcs/0007-flow-modules.md)

## RFC Goals

* Explain the use cases of modules
* Explain what modules are
* Go over possible syntax for modules
* Go over pros and cons of modules

## Summary

One of the primary goals for the production usage of Agent Flow is parity with the static subsystem. One of the features of the static subsystem is [scraping service](https://github.com/grafana/agent/blob/main/docs/sources/configuration/scraping-service.md). Scraping service allows a user to run a fleet of agents and have thousands of scrape configurations distributed among running Agents. Through discussions within the Agent Team, we did not want to limit dynamically loading content to only scrape configs but allow most components to be loaded and used.

During this time the Agent team saw a lot of potential in the idea of "modules." Modules package up sets of components for specific workflows, publishing common use cases and allowing better usage for internal customers in the Agent as a Service model.

## Goals

* Support single module loading via `module.string`
* Enable re-use of common patterns
* Allow loading a module from a string
* Allow modules to load other modules
* Modules should be sandboxed except via arguments and exports

### Enable re-use of common patterns

Common functionality can be wrapped in a set of common components that form a module. These shared modules can then be used instead of reinventing use cases.

### Allow loading a module from a string

Modules will not care about the source of a string. In the case of a `module.string` the module will take in one string of valid river configuration.

### Allow modules to load other modules

Modules will be able to load other modules, with reasonable safe guards. There will be a stack limit for the depth of sub-modules and modules will be unable to load themselves.

### Modules should be sandboxed except via arguments and exports

Modules cannot directly access children or parent modules except through predefined arguments and exports.

## Non Goals

Non goals represent capabilities that are not going to be done in the initial release of modules but may come in later versions.

* Add additional capabilities to load strings
* Any type of versioning
* Any user interface work beyond ensuring it works as the UI currently does
* Any sort of security for modules and what modules are allowed
* Any sort of metadata



### Add additional capabilities to load strings

Modules will not care about the source of the string that represents the river syntax, nor will modules have any inherent reload semantics. The component supplying the string will be responsible for the source and will notify the module when the input changes so that it can utilize the new river configuration.

### Any type of versioning

Modules will not contain any sort of versioning nor will check for compatibility outside the normal river checks.

### Any user interface work beyond ensuring it works as the UI currently does

Users will not be able to drill into modules, they will be represented as any other normal component.

## Example

```river
// module
argument "password" {
    optional = false
    comment = "password for mysql"
}

argument "username" {
    optional = false
    comment = "username for mysql"
}

export "targets" {
    comment = "targets for the integration"
    value = integrations.mysql.server1.targets
}

integrations.mysql "server1" {
    username = argument.username.value
    password = argument.password.value
}

```

```river
// parent

local.file "mysql" {
    filename = "/test/mysql.river"
}

module.string "mysql" {
    content = local.file.mysql.content
    arguments = {
        {
            "password" = PASSWORD,
            "username" = USERNAME,
        }
    }
}

prometheus.scrape "scraper" {
    targets = module.string.mysql.exports.targets
}

```

## Limitations

* A module cannot directly or indirectly load itself, this will not be enforced by the system
* Singleton components are not supported at this time. Example [node_exporter](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.integration.node_exporter/).
* Modules will not prevent competing resources, such as starting a server on the same port
* [Configuration blocks](https://grafana.com/docs/agent/latest/flow/reference/config-blocks/#configuration-blocks) will not be supported.
* Names of arguments and exports within a module must be unique across that module.

## Proposal

Add the ability to load `modules` as subgraphs to the primary `graph`. Modules may call other modules within a reasonable stack size depth. Modules are represented as a river string that is interpreted with a defined set of arguments and exports.

The initial component will be `module.string` that will load a single module. Internally these modules will be namespaced, so they cannot affect children or parent graphs except via arguments and exports.

Modules will have access to any standard function and any other component exempting singletons. Internally each component in the module will have an `id` that is prepended with the parent's `id` for identification purposes outside of the module. Within the module a component can reference another sibling component normally. There are no known limits on the datatype that a module can use as an argument or export.

`modules.multiple` will use the key of the `map(string)` to uniquely identify a specific module.

### Component Options


Given the above example, the `id` of `integrations.mysql "server1"` would be `module.string.mysql.integrations.mysql.server1`. The `data-agent` field would also be prefixed. There are some inherent issues, deeply nested metrics are likely to run into Prometheus label value limits. On Windows platforms there could be issues with the `data-agent` length. These are issues that currently exist in Agent Flow but are more easily hit using deeply nested modules.


### Failure Modes

#### Option 1: When a module fails then fail itself and children.

If an error occurs while re-evaluating a module then the module marks itself as unhealthy and unloads the original module and all children.

*Pros*

* Simple to implement
* Easy to understand

*Cons*

* One failure mode can cascade

#### Option 2: Modules keep last good value

If an error occurs while re-evaluating a module then the module marks itself as unhealthy and attempts to keep the original module. This may have an issue with cascading failures, if a module depends on a module then the system may enter an inconsistent state while applying and then rolling back the change.

For example, `Module A` has two sub-modules `Module B` and `Module C`. During reevaluation `Module B` reloads appropriately but `Module C` fails. `Module A` unloads both modules and then reloads the last good string. In the case that the last good string also fails then `Module A` is unhealthy and non-functional and `Module A's` submodules do not exist.

*Pros*

* Allows more resilient usage

*Cons*

* Can create undefined behavior
* Complex to unload and reload

# Possible Example Documentation for `argument`

## Arguments

The following arguments are supported:

Name            | Type                | Description                                                                                | Default | Required
--------------- | ------------------- | ------------------------------------------------------------------------------------------ |---------| --------
`optional`  | `bool` | If an argument has to be specified. |    "false"     | no
`comment`  | `string` | Comment describing what the argument is used for |    ""     | no
`default`  | `any` | Default value if unspecified |         | no

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`value` | `any` | The represented value of the argument.


# Example Documentation for `export`

## Arguments

The following arguments are supported:

Name            | Type                | Description                                                                                | Default | Required
--------------- | ------------------- | ------------------------------------------------------------------------------------------ |---------| --------
`comment`  | `string` | Comment describing what the export is used for |    ""     | no

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`value` | `any` | The represented value of the export.


# Example Documentation for `module.string`

## Arguments

The following arguments are supported:

Name            | Type            | Description                                                                                                                   | Default | Required
--------------- | --------------- |-------------------------------------------------------------------------------------------------------------------------------|---------| --------
`arguments`  | `map(string)` | Map of items to pass to module. It is possible to include arguments that are not needed. Any required arguments are required. |    "'{}'"     | no
`content`  | `string` | River configuration to be loaded.                                                                                             |    "''"     | yes

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`exports` | `map(string)` | The set of exports where the key is the name of an export and the value is it's value
