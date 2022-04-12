# Flow expressions component anatomy

* Date: 2022-04-12
* Author: Robert Fratto (@rfratto)
* PR: [grafana/agent#1559](https://github.com/grafana/agent/pull/1599)
* Status: Draft

## Background

[RFC-0004][] described the general need for Grafana Agent to support
component-based system to make Grafana Agent easier to use and develop called
"Grafana Agent Flow."

This document will outline in further detail what a component is composed of
and how components interact with one another. This document focuses solely on
the expressions-based architecture. Detailing how components would work in a
messages-based architecture (as described by [RFC-0004][]) is out of scope.

> **NOTE**: This document uses HCL as the configuration language for
> components, selected for its native support for expressions. The actual
> implementation of Grafana Agent Flow can support a non-HCL language, but an
> HCL-based language is the best fit given the proposed design of how
> components interact.
>
> A complete specification of the HCL language can be found in the
> [HCL repository][hcl-spec].

## Goals

* Describe the input and output of components
* Describe how component expose health for user introspection
* Describe how components are named
* Describe how two components may interact

## Non-Goals

* List components that will be implemented
* Explore the implementation details of specific components
* Describe the anatomy of a messages-based component: this document focuses on
  expressions, and another proposal would be needed for messages.

## Proposal

Components are intended to be standalone units of logic which play some role in
the overall telemetry pipeline. Example components may be:

* A component which performs service discovery (e.g., Prometheus' Kubernetes SD).
* A component which scrapes metrics from a service discovery target.
* A component which writes scraped metrics to a series of endpoints.
* A component which retrieve API keys from an external store such as Vault.

### Component structure

This section will use a Kubernetes SD component for an example.

A component is composed of:

1. User-supplied input
2. Output state
3. Component status

#### Input and Output

The user-supplied input holds settings to configure the behavior of the input.
These are supplied by an HCL block which describes the component, such as:

```hcl
discovery "kubernetes" "pods" {
  role = "pod"
}
```

Components may choose to expose parts of their state as output. This example
component would likely include the current set of discovered pods as its
output. The output state is a set of key-value pairs, where the `targets` key
would be a list of discovered targets for the component.

Component output is declarative, describing the current state of the component.
This allows the values for both user-supplied input and component output state
to be referenced by other components:

```hcl
discovery "kubernetes" "pods" {
  role = "pod"

  // Output state:
  // - targets: list of discovered pod targets
}

discovery "kubernetes" "nodes" {
  role = "node"

  // Output state:
  // - targets: list of discovered nodes targets
}

prometheus_scrape "kubernetes-resources" {
  // Scrape Prometheus metrics from the combination (concatenation) of
  // discovered Kubernetes pods and nodes.
  //
  // This example only references output state, but
  // discovery.kubernetes.pods.role can also be used within expressions for the
  // input of other components.
  targets = concat(
    discovery.kubernetes.pods.targets,
    discovery.kubernetes.nodes.targets,
  )

  // Output state: <none>
}
```

#### Component updating

Components will be updated if any component they reference has changed its
output state. Similarly, a component will emit an event to the Grafana Agent
Flow system when its output state has changed.

Components should not be directly aware of this relationship: rather, they will
receive their new input so they can update their internal state, and emit a
"state changed" event if there is new output state available.

#### Component status

The health of a component should be observable by a user without having that
health by exposed as output state.

Component status is a debug-only set of key-value pairs which describe the
current status of a component. The status of a component may not be referenced
by other components, and may only be viewed by a user when specifically
requested, such as when viewing an HTTP `/status` endpoint. No events are
emitted by a component when its status changes.

The `prometheus_scrape` component may include status denoting the last scrape
time, duration, and scrape errors for each of its input targets.

### Component naming and identification

Components are identified by three parameters:

* (Optional) A component namespace, where a namespace is used for a collection
  of related components. A component without a namespace is said to occupy the
  global namespace.
* (Required) A component name, which uniquely identifies that component in its
  namespace.
* (Optional) A user-supplied component identifier, which uniquely identifies
  the specific instance of the component in the configuration file.

The fully-qualified identifier of a component is the combination of the used
identifier parameters separated by `.`. For example, `remote.vault.api-key` is
composed of:

* A namespace of `remote`
* A name of `vault`
* A user-supplied idenfifier of `api-key`

The fully-qualified identifier for a component must be globally unique across
the process. This implies that components which support user-supplied
identifiers can be defined multiple times; there may be many `remote.vault.*`
components.

Components which do not have a namespace or user-supplied identifier are
referred to only by name (e.g., `node_exporter`). The lack of a user-supplied
identifier implies that there may be no more than a single `node_exporter`
component defined by the user.

### Component definition

This document defines components as HCL blocks, where:

* If the component has a namespace:
  * The block type is the component namespace
  * The first label is the component name
  * The second label is the user-supplied identifier (if supported by the component)
* If the component does not have a namespace:
  * The block type is the component name
  * The second label is the user-supplied identifier (if supported by the component)

```hcl
# namespace: remote
# name:      vault
# user ID:   api-key
remote "vault" "api-key" {
  # ... component settings ...
}

# namespace: <none>
# name:      node_exporter
# user ID:   <none>
node_exporter {
  # ... component settings ...
}

# namespace: <none>
# name:    remote_write
# user ID: default
remote_write "default" {
  # ... component settings ...
}
```

### Concerns

#### `for_each`

Users of HCL may opt-in to a `for_each` attribute which expands one HCL block
into a dynamic set. For example:

```terraform
resource "azurerm_resource_group" "rg" {
  for_each = {
    a_group = "eastus"
    another_group = "westus2"
  }
  name     = each.key
  location = each.value
}
```

It is likely that Grafana Agent Flow will want to support a similar concept,
such as for dynamically running integrations based on a set of discovered
targets:

```hcl
discovery "kubernetes" "redis-pods" {
  role = "pod"
  # ... filter for Redis pods ...
}

integration "redis" "kubernetes" {
  for_each = discovery.kubernetes.redis-pods.targets
  # ... redis_exporter settings ...
}
```

While users are still required for the `integration.redis.kubernetes`
fully-qualified identifier to be globally unique, there would be many
components running with that identifier.

In this specific scenario, this turns `integration.redis.kubernetes` into an
array of components: `integration.redis.kubernetes[0]`,
`integration.redis.kubernetes[1]` and so on.

[RFC-0004]: https://github.com/grafana/agent/pull/1546
[hcl-spec]: https://github.com/hashicorp/hcl/blob/main/hclsyntax/spec.md
