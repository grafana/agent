---
canonical: https://grafana.com/docs/agent/latest/flow/
title: Flow mode
weight: 400
---

# Flow mode

The Flow mode of Grafana Agent (also called Grafana Agent Flow) is a
_component-based_ revision of Grafana Agent with a focus on ease-of-use,
debuggability, and ability to adapt to the needs of power users.

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

## Online configurator

Check out the [Grafana Agent Configuration Generator](https://grafana.github.io/agent-configurator/) to get a head start on creating
flow code. Note that this is experimental and not all components are supported.

## Next steps

* [Install][] Grafana Agent in flow mode.
* Learn about the core [Concepts][] of flow mode.
* Follow our [Getting started][] guides for Grafana Agent in flow mode.
* Follow our [Tutorials][] to get started with Grafana Agent in flow mode.
* Learn how to use the [Configuration language][].
* Check out our [Reference][] documentation to find specific information you
  might be looking for.

[Install]: {{< relref "./setup/install/" >}}
[Concepts]: {{< relref "./concepts/" >}}
[Getting started]: {{< relref "./getting-started/" >}}
[Tutorials]: {{< relref "./tutorials/ ">}}
[Configuration language]: {{< relref "./config-language/" >}}
[Reference]: {{< relref "./reference" >}}

