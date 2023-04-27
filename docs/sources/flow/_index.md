---
title: Grafana Agent Flow
weight: 900
---

# Grafana Agent Flow (Experimental)

Grafana Agent Flow is a _component-based_ experimental revision of Grafana
Agent with a focus on ease-of-use, debuggability, and ability to adapt to the
needs of power users.

Components allow for reusability, composability, and focus on a single task. 

* **Reusability** allows for the output of components to be reused as the input for multiple other components.
* **Composability** allows for components to be chained together to form a pipeline.
* **Single task** means the scope of a component is limited to one narrow task and thus has fewer side effects.

> **EXPERIMENTAL**: Grafana Agent Flow is an [experimental][] feature.
> Experimental features are subject to frequent breaking changes and are
> subject for removal if the experiment doesn't work out.
>
> You should only use Grafana Agent Flow if you are okay with bleeding edge
> functionality and want to provide feedback to the developers. It is not
> recommended to use Grafana Agent Flow in production.

[experimental]: {{< relref "../operation-guide#stability" >}}

## Features

* Write declarative configurations with a Terraform-inspired configuration
  language.
* Declare components to configure parts of a pipeline.
* Use expressions to bind components together to build a programmable pipeline.
* Includes a UI for debugging the state of a pipeline.

## Example

```river
// Discover Kubernetes pods to collect metrics from.
discovery.kubernetes "pods" {
  role = "pod"
}

// Scrape metrics from Kubernetes pods and send to a prometheus.remote_write
// component.
prometheus.scrape "default" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [prometheus.remote_write.default.receiver]
}

// Get an API key from disk.
local.file "apikey" {
  filename  = "/var/data/my-api-key.txt"
  is_secret = true
}

// Collect and send metrics to a Prometheus remote_write endpoint.
prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9009/api/prom/push"

    http_client_config {
      basic_auth {
        username = "MY_USERNAME"
        password = local.file.apikey.content
      }
    }
  }
}
```

## Next steps

* Learn about the [core concepts][] of Grafana Agent Flow.
* Follow our [tutorials][] to get started with Grafana Agent Flow.
* Learn how to use Grafana Agent Flow's [configuration language][].
* Check out our [reference documentation][] to find specific information you
  might be looking for.

[core concepts]: {{< relref "./concepts/" >}}
[tutorials]: {{< relref "./tutorials/ ">}}
[configuration language]: {{< relref "./config-language/" >}}
[reference documentation]: {{< relref "./reference" >}}

## Current limitations

The goal of Grafana Agent Flow is to eventually support the same use cases that
Grafana Agent does today. Some functionality may be missing while Grafana Agent
Flow is still in development:

* Logging-specific components
* Tracing-specific components
* An equivalent list of integrations
* An equivalent to the scraping service

## Provide feedback

Feedback about Grafana Agent Flow and its configuration language can be
provided in our dedicated [GitHub discussion for feedback][feedback].

[feedback]: https://github.com/grafana/agent/discussions/1969
