---
aliases:
- /docs/agent/latest/concepts/component-controller
title: Component controller
weight: 200
---

# Component controller

The _component controller_ is the core part of Grafana Agent Flow which manages
components at runtime.

It is responsible for:

* Reading and validating the config file
* Managing the lifecycle of defined components
* Evaluating the arguments used to configure components
* Reporting the health of defined components

## Component graph

As discussed in [Components][], a relationship between components is created
when an expression is used to set the argument of one component to an export of
another component.

The set of all components and the relationships between them define a [directed
acyclic graph][DAG] (DAG), which informs the component controller which
references are valid and in what order components must be evaluated.

For a config file to be valid, components must not reference themselves or
contain a cyclic reference:

```river
// INVALID: local.file.some_file may not reference itself:
local.file "self_reference" {
  filename = local.file.self_reference.content
}
```

```river
// INVALID: cyclic reference between local.file.a and local.file.b:
local.file "a" {
  filename = local.file.b.content
}
local.file "b" {
  filename = local.file.a.content
}
```

## Component evaluation

A component is _evaluated_ when its expressions are computed into concrete
values and used to configure the component. The component controller is
finished loading once all components have been evaluated and are running.

The component controller is guaranteed to only evaluate a given component after
evaluating all of the components the given component references. Components
which do not reference other components may be evaluated at any time during the
evaluation process.

## Component reevaluation

As mentioned in [Components][], a component is dynamic: a component may update
its exports any number of times throughout its lifetime.

When a component updates its exports, a _controller reevaluation_ is triggered:
the component controller will reevaluate any component which references the
changed component, any components which reference those components, and so on,
until all affected components have been reevaluated.

## Component health

At any given time, a component may have one of the following health states:

1. Unknown: default state, the component isn't running yet
2. Healthy: the component is working as expected
3. Unhealthy: the component is not working as expected
4. Exited: the component has stopped and is no longer running

By default, the component controller determines the health of a component. The
component controller will mark a component has healthy as long as that
component is running and its most recent evaluation was successful.

Some components may report their own component-specific health information. For
example, the `local.file` component will report itself unhealthy if the file it
was watching gets deleted.

The overall health of a component is reported by combining the
controller-reported health of the component with the component-specific health
information.

An individual component's health is independent from the health of any other
components it references: a component may be marked as healthy even if it
references the exports of an unhealthy component.

## Handling evaluation failures

When a component fails to evaluate, it is marked as unhealthy with the reason
for why the evaluation failed.

When an evaluation fails, the component continues operating as normal: it
continues using its previous set of evaluated arguments, and it may continue
exporting new values.

This prevents failure propagation: if your `local.file` component which watches
API keys suddenly stops working, other components will continue using the last
valid API key until the component returns to a healthy state.

## Updating the config file

Both the `/-/reload` HTTP endpoint and the `SIGHUP` signal can be used to
inform the component controller to reload the config file. When this happens,
the component controller will synchronize the set of running components with
the ones in the config file, removing components which are no longer defined in
the config file and creating new components which were added to the config
file. All components managed by the controller will be reevaluated after
reloading.

[Components]: {{< relref "./components.md" >})
[DAG]: https://en.wikipedia.org/wiki/Directed_acyclic_graph
