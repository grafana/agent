---
aliases:
- /docs/agent/latest/flow/configuration-language/components
title: Components
weight: 300
---

# Components
Components are the defining feature of Grafana Agent Flow. They are autonomous
pieces of business logic that perform a single task (like retrieving secrets or
collecting Prometheus metrics) and are wired together to form programmable
pipelines of telemetry data.

Under the hood, components are orchestrated via the [_component
controller_]({{< relref "../concepts/component_controller.md" >}}), who is
responsible for scheduling them, reporting their health and debug status, and
continuously re-evaluating their inputs and outputs.

Each top-level River _block_ will instantiate a new component. A component
consists of its name (which describes what the component is responsible for) as
well as an optional user-specified label.

You can see a list of all available components [here]({{< relref
"../components/_index.md" >}}). Each component features a complete reference
pages, so getting a component to work should be as easy as reading its
documentation and copy/pasting from an example.

## Arguments and Outputs
Most user interactions with components will center around two basic concepts;
_Arguments_ and _Exports_.

* _Arguments_ work much like function arguments in your favorite programming
 language. They can be any number of attributes or nested blocks and they're
used to configure the component's behavior. Most components will feature both
required and optional arguments. Any arguments that are not provided will take
on their default values.

* _Exports_ are zero or more named return values, that can be referred to by
 other components. There is no limit to what an export might be; it can be any
River value, or Go type (eg. channels, interfaces).

Here's a quick example; the following block defines a `local.file` component
labelled "targets". The `local.file.targets` component will then expose the
file `content` in its Exports.

The `filename` attribute is a _required_ argument; the user can also define a
number of _optional_ ones (in this case `is_sensitive`, `detector` and
`poll_frequency`), which configure how and how often the file should be polled
as well as whether its contents are safe to be presented as plaintext back to
the user.

```river
local.file "targets" {
	// Mandatory Argument
	filename = "/etc/agent/targets" 

	/* Optional Arguments
	is_sensitive   = <boolean>
	detector       = < "fsnotify" | "poll" >
	poll_frequency = <duration> 
	*/

	// Export: a field named `content`
	// It can be referred to as `local.file.token.content`
}
```

To wire components together, you just use the Exports of one as the Arguments
of another.

For example, here's a component that scrapes Prometheus metrics. The `targets`
field is populated with two scrape targets; a constant one `localhost:9001` and
an expression that ties the target to the contents of the
`local.file.target.content`.

```river
prometheus.scrape "default" {
	targets = [
		{ "__address__" = local.file.target.content }, // tada!
		{ "__address__" = "localhost:9001" },
	] 

	forward_to = [prometheus.remote_write.default.receiver]
	scrape_config {
		job_name = "default"
	}
}
```

All Arguments and Exports have an underlying [type]({{< relref
"./expressions/types_and_values.md" >}}). River will type-check expressions
before assigning a value to an attribute; the documentation of each component
will have more information about the ways that you can wire components
together.

River attributes that appear as Arguments and Exports can carry around more
than the basic types you'd expect such as strings, numbers or booleans; they
are able to wrap any Go value (eg. channels or interfaces), which makes for
powerful tooling when building more complex pipelines.

## DAG
The relationships between components form a [Directed Acyclic Graph](https://en.wikipedia.org/wiki/Directed_acyclic_graph).

This graph is used by River to resolve dependencies between components, set the correct order for component evaluation, and ensure the validity of references. River does not allow components that reference themselves, or other relationship that end up in cyclical references, or invalid references.

In the previous example, the contents of the `local.file.target.content` expression must first be evaluated in a concrete value then type-checked and substituted into `prometheus.scrape.default` for it to be configured in turn.

## Error reporting
Like in the case of syntax errors, River will also enrich component-related
errors with useful information to help with debugging.

```
Error: ./cmd/agent/example-config.river:16:21: component "local.file.this_files.content" does not exist

15 |         "__address__"   = "localhost:12345",
16 |         "dynamic_label" = local.file.this_files.content,
   |                           ^^^^^^^^^^^^^^^^^^^^^^^
17 |     }]

Error: ./cmd/agent/example-config.river:25:1: Failed to build component: building component: duplicate remote write configs are not allowed, found duplicate for URL: http://localhost:9009/api/prom/push

24 |
25 | | prometheus.remote_write "default" {
   |  _^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
26 | |     remote_write {
27 | |         url = "http://localhost:9009/api/prom/push"
28 | |     }
29 | |     remote_write {
30 | |         url = "http://localhost:9009/api/prom/push"
31 | |     }
32 | | }
   | |_^
33 |
```
