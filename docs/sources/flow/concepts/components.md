---
aliases:
- ../../concepts/components/
canonical: https://grafana.com/docs/agent/latest/flow/concepts/components/
title: Components
weight: 100
---

# Components

_Components_ are the building blocks of Grafana Agent Flow. Each component is
responsible for handling a single task, such as retrieving secrets or
collecting Prometheus metrics.

Components are composed of two parts:

* Arguments: settings which configure a component.
* Exports: named values which a component exposes to other components.

Each component has a name which describes what that component is responsible
for. For example, the `local.file` component is responsible for retrieving the
contents of files on disk.

Components are specified in the config file by first providing the component's
name with a user-specified label, and then by providing arguments to configure
the component:

```river
discovery.kubernetes "pods" {
  role = "pod"
}

discovery.kubernetes "nodes" {
  role = "node"
}
```

> Components are referenced by combining the component name with its label. For
> example, a `local.file` component labeled `foo` would be referenced as
> `local.file.foo`.
>
> The combination of a component's name and its label must be unique within the
> configuration file. This means multiple instances of a component may be
> defined as long as each instance has a different label value.

## Pipelines

Most arguments for a component in a config file are constant values, such
setting a `log_level` attribute to the quoted string `"debug"`:

```river
log_level = "debug"
```

_Expressions_ can be used to dynamically compute the value of an argument at
runtime. Among other things, expressions can be used to retrieve the value of
an environment variable (`log_level = env("LOG_LEVEL")`) or to reference an
exported field of another component (`log_level = local.file.log_level.content`).

When a component's argument references an exported field of another component,
a dependant relationship is created: a component's input (arguments) now
depends on another component's output (exports). The input of the component
will now be re-evaluated any time the exports of the components it references
get updated.

The flow of data through the set of references between components forms a
_pipeline_.

An example pipeline may look like this:

1. A `local.file` component watches a file on disk containing an API key.
2. A `prometheus.remote_write` component is configured to receive metrics and
   forward them to an external database using the API key from the `local.file`
   for authentication.
3. A `discovery.kubernetes` component discovers and exports Kubernetes Pods
   where metrics can be collected.
4. A `prometheus.scrape` component references the exports of the previous
   component, and sends collected metrics to the `prometheus.remote_write`
   component.

<p align="center">
<img src="../../../assets/concepts_example_pipeline.svg" alt="Flow of example pipeline" width="500" />
</p>

The following config file represents the above pipeline:

```river
// Get our API key from disk.
//
// This component has an exported field called "content", holding the content
// of the file.
//
// local.file.api_key will watch the file and update its exports any time the
// file changes.
local.file "api_key" {
  filename  = "/var/data/secrets/api-key"

  // Mark this file as sensitive to prevent its value from being shown in the
  // UI.
  is_secret = true
}

// Create a prometheus.remote_write component which other components can send
// metrics to.
//
// This component exports a "receiver" value which can be used by other
// components to send metrics.
prometheus.remote_write "prod" {
  endpoint {
    url = "https://prod:9090/api/v1/write"

    basic_auth {
      username = "admin"

      // Use our password file for authenticating with the production database.
      password = local.file.api_key.content
    }
  }
}

// Find Kubernetes pods where we can collect metrics.
//
// This component exports a "targets" value which contains the list of
// discovered pods.
discovery.kubernetes "pods" {
  role = "pod"
}

// Collect metrics from Kubernetes pods and send them to prod.
prometheus.scrape "default" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [prometheus.remote_write.prod.receiver]
}
```
