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


## Configuring components
Each top-level River _block_ will instantiate a new component. All components
are identified by their name, describing what the component is responsible for,
while some allow or require to provide an extra user-specified _label_.

You can see a list of all available components [here]({{< relref "../components/_index.md" >}}).
Each one features a complete reference page, so getting a component to work for
you should be as easy as reading its documentation and copy/pasting from an
example.

## Arguments and Exports
Most user interactions with components will center around two basic concepts;
_Arguments_ and _Exports_.

* _Arguments_ are settings which modify the behavior of a component. They can
 be any number of attributes or nested unlabeled blocks, some of them being
required for the component to work and some being optional. Any optional
arguments that are not overriden, will take on their default values.

* _Exports_ are zero or more named return values that can be referred to by
 other components and can be any River value.

Here's a quick example; the following block defines a `local.file` component
labelled "targets". The `local.file.targets` component will then expose the
file `content` as a string in its Exports.

The `filename` attribute is a _required_ argument; the user can also define a
number of _optional_ ones, in this case `detector`, `poll_frequency` and
`is_sensitive`, which configure how and how often the file should be polled
as well as whether its contents are safe to be presented as plaintext back to
the user.

```river
local.file "targets" {
	// Required Argument
	filename = "/etc/agent/targets" 

	// Optional Arguments
	//   is_sensitive   = <boolean>
	//   detector       = < "fsnotify" | "poll" >
	//   poll_frequency = <duration> 

	// Exports: a single field named `content`
	// It can be referred to as `local.file.token.content`
}
```

## Referencing components
To wire components together, one can use the Exports of one as the Arguments
to another.

For example, here's a component that scrapes Prometheus metrics. The `targets`
field is populated with two scrape targets; a constant one `localhost:9001` and
an expression that ties the target to the contents of
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

Every time the file contents change and the `local.file` component re-evaluates
its Exports, the value will propagated to `prometheus.scrape` so it can start
scraping the new target.

All Arguments and Exports have an underlying [type]({{< relref "./expressions/types_and_values.md" >}}).
River will type-check expressions before assigning a value to an attribute; the
documentation of each component will have more information about the ways that
you can wire components together.

In the previous example, the contents of the `local.file.target.content`
expression must first be evaluated in a concrete value then type-checked and
substituted into `prometheus.scrape.default` for it to be configured in turn.

