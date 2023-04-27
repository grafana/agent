---
aliases:
- ../../concepts/component-controller/
title: Component controller
weight: 200
---

# Component controller

The _component controller_ is the core part of Grafana Agent Flow which manages
components at runtime.

It is responsible for:

* Reading and validating the config file.
* Managing the lifecycle of defined components.
* Evaluating the arguments used to configure components.
* Reporting the health of defined components.

## Component graph

As discussed in [Components][], a relationship between components is created
when an expression is used to set the argument of one component to an exported
field of another component.

The set of all components and the relationships between them define a [directed
acyclic graph][DAG] (DAG), which informs the component controller which
references are valid and in what order components must be evaluated.

For a config file to be valid, components must not reference themselves or
contain a cyclic reference:

```river
// INVALID: local.file.some_file can not reference itself:
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
values. The computed values are then used to configure the component's runtime
behavior. The component controller is finished loading once all components are
evaluated, configured, and running.

The component controller only evaluates a given component after evaluating all
of that component's dependencies. Component that do not depend on other
components can be evaluated at any time during the evaluation process.

## Component reevaluation

As mentioned in [Components][], a component is dynamic: a component can update
its exports any number of times throughout its lifetime.

When a component updates its exports, a _controller reevaluation_ is triggered:
the component controller reevaluates any component which references the changed
component, any components which reference those components, and so on, until
all affected components are reevaluated.

## Component health

At any given time, a component can have one of the following health states:

1. Unknown: default state, the component isn't running yet.
2. Healthy: the component is working as expected.
3. Unhealthy: the component is not working as expected.
4. Exited: the component has stopped and is no longer running.

By default, the component controller determines the health of a component. The
component controller marks a component as healthy as long as that component is
running and its most recent evaluation succeeded.

Some components can report their own component-specific health information. For
example, the `local.file` component reports itself as unhealthy if the file it
was watching gets deleted.

The overall health of a component is determined by combining the
controller-reported health of the component with the component-specific health
information.

An individual component's health is independent from the health of any other
components it references: a component can be marked as healthy even if it
references an exported field of an unhealthy component.

## Handling evaluation failures

When a component fails to evaluate, it is marked as unhealthy with the reason
for why the evaluation failed.

When an evaluation fails, the component continue operating as normal: it
continues using its previous set of evaluated arguments, and it can continue
exporting new values.

This prevents failure propagation: if your `local.file` component which watches
API keys suddenly stops working, other components continues using the last
valid API key until the component returns to a healthy state.

## Updating the config file

Both the `/-/reload` HTTP endpoint and the `SIGHUP` signal can be used to
inform the component controller to reload the config file. When this happens,
the component controller will synchronize the set of running components with
the ones in the config file, removing components which are no longer defined in
the config file and creating new components which were added to the config
file. All components managed by the controller will be reevaluated after
reloading.

[Components]: {{< relref "./components.md" >}}
[DAG]: https://en.wikipedia.org/wiki/Directed_acyclic_graph
