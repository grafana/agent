---
aliases:
- /docs/grafana-cloud/agent/flow/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/
- /docs/grafana-cloud/send-data/agent/flow/
canonical: https://grafana.com/docs/agent/latest/flow/
description: Grafana Agent Flow is a component-based revision of Grafana Agent with
  a focus on ease-of-use, debuggability, and adaptability
title: Flow mode
weight: 400
cascade:
  PRODUCT_NAME: Grafana Agent Flow
  PRODUCT_ROOT_NAME: Grafana Agent
---

# {{% param "PRODUCT_NAME" %}}

{{< param "PRODUCT_NAME" >}} is a _component-based_ revision of {{< param "PRODUCT_ROOT_NAME" >}} with a focus on ease-of-use,
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
// Discover Kubernetes pods to collect metrics from
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


## {{% param "PRODUCT_NAME" %}} configuration generator

The {{< param "PRODUCT_NAME" >}} [configuration generator](https://grafana.github.io/agent-configurator/) will help you get a head start on creating flow code.

{{< admonition type="note" >}}
This feature is experimental, and it doesn't support all River components.
{{< /admonition >}}

## Next steps

* [Install][] {{< param "PRODUCT_NAME" >}}.
* Learn about the core [Concepts][] of {{< param "PRODUCT_NAME" >}}.
* Follow our [Tutorials][] for hands-on learning of {{< param "PRODUCT_NAME" >}}.
* Consult our [Tasks][] instructions to accomplish common objectives with {{< param "PRODUCT_NAME" >}}.
* Check out our [Reference][] documentation to find specific information you
  might be looking for.

[Install]: {{< relref "./get-started/install/" >}}
[Concepts]: {{< relref "./concepts/" >}}
[Tasks]: {{< relref "./tasks/" >}}
[Tutorials]: {{< relref "./tutorials/ ">}}
[Reference]: {{< relref "./reference" >}}

