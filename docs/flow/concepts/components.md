---
aliases:
- /docs/agent/latest/concepts/components
title: Components
weight: 100
---

# Components

_Components_ are the building blocks of Grafana Agent Flow. Each component is
responsible for handling a single task, such as retrieving secrets or
collecting Prometheus metrics.

Components are composed of two parts:

* Arguments: settings which configure a component
* Exports: named values which a component exposes to other components

Each component has a name which describes what that component is responsible
for. For example, the `local.file` component is responsible for retrieving the
contents of files on disk.

Components are specified by users by first providing the component's name with
a user-specified label, and then by providing arguments to configure the
component.

> Components are referenced by combining the component name with its label. For
> example, a `local.file` component labeled `foo` would be referenced as
> `local.file.foo`.
>
> The combination of a component's name and its label must be unique within the
> configuration file.

## Pipelines

Most arguments a user specifies for a component will be some constant value,
such setting a `log_level` attribute to the quoted string `"debug"` (`log_level
= "debug"`).

Users may also use _expressions_ to dynamically compute the value of an
argument at runtime. Among other things, expressions may be used to retrieving
the value of an environment variable (`log_level = env("LOG_LEVEL")`) or
referencing the export of another component (`log_level =
local.file.log_level.contents`).

When a user configures a component's argument to reference an export of another
component, a relationship is created: a component's input (arguments) now
depends on another component's output (exports). The input of the component
will now be re-evaluated any time the exports of the components it references
get updated.

The flow of data through the set of references between components forms a
_pipeline_.

An example pipeline may look like this:

1. A `local.file` component watches a file on disk containing an API key.
1. A `metrics.remote_write` component is configured to receive metrics and
   forward them to an external database using the API key from the `local.file`
   for authentication.
2. A `discovery.kubernetes` component discovers and exports Kubernetes Pods
   where metrics can be collected.
3. A `metrics.scrape` component references the exports of the previous
   component, and sends collected metrics to the `metrics.remote_write`
   component.

A user would use this config file to represent the above pipeline:

```river
// Get our API key from disk.
//
// This component has an exported value called "contents", holding the contents
// of the file.
//
// local.file.api_key will watch the file and update its exports any time the
// file changes.
local.file "api_key" {
  filename  = "/var/data/secrets/api-key"

  // Mark this file as sensitive to prevent its value from being shown to
  // users.
  is_secret = true
}

// Create a metrics.remote_write component which other components can send
// metrics to.
//
// This component exports a "receiver" value which can be used by other
// components to send metrics.
metrics.remote_write "prod" {
  remote_write {
    url = "https://prod:9090/api/v1/write"

    basic_auth {
      username = "admin"

      // Use our password file for authenticating with the production database.
      password = local.file.api_key.contents
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
metrics.scrape "default" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [metrics.remote_write.prod.receiver]
}
```
