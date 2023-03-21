---
title: Grafana Agent Flow
weight: 900
---

# Grafana Agent Flow

Grafana Agent Flow is a _component-based_ revision of Grafana Agent with a
focus on ease-of-use, debuggability, and ability to adapt to the needs of power
users.

Components allow for reusability, composability, and focus on a single task.

* **Reusability** allows for the output of components to be reused as the input for multiple other components.
* **Composability** allows for components to be chained together to form a pipeline.
* **Single task** means the scope of a component is limited to one narrow task and thus has fewer side effects.

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

    basic_auth {
      username = "MY_USERNAME"
      password = local.file.apikey.content
    }
  }
}
```

## Next steps

* Learn about the [core concepts][] of Grafana Agent Flow.
* Follow our Grafana Agent Flow [Getting Started][] guides.
* Follow our [tutorials][] to get started with Grafana Agent Flow.
* Learn how to use Grafana Agent Flow's [configuration language][].
* Check out our [reference documentation][] to find specific information you
  might be looking for.

[core concepts]: {{< relref "./concepts/" >}}
[Getting Started]: {{< relref "./getting-started/" >}}
[tutorials]: {{< relref "./tutorials/ ">}}
[configuration language]: {{< relref "./config-language/" >}}
[reference documentation]: {{< relref "./reference" >}}

