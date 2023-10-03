---
aliases:
- ../../configuration-language/expressions/referencing-exports/
- /docs/grafana-cloud/agent/flow/config-language/expressions/referencing_exports/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/config-language/expressions/referencing_exports/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/config-language/expressions/referencing_exports/
canonical: https://grafana.com/docs/agent/latest/flow/config-language/expressions/referencing_exports/
title: Referencing component exports
description: Learn about referencing component exports
weight: 200
---

# Referencing component exports
Referencing exports is what enables River to dynamically configure and connect
components using expressions. While components can work in isolation, they're
more useful when one component's behavior and data flow is bound to the exports
of another, building a dependency relationship between the two.

Such references can only appear as part of another component's arguments or a
config block's fields. That means that components cannot reference themselves.

## Using references
These references are built by combining the component's name, label and named
export with dots.

For example, the contents of a file exported by the `local.file` component
labeled `target` might be referenced as `local.file.target.content`.
Similarly, a `prometheus.remote_write` component instance labeled `onprem` will
expose its receiver for metrics on `prometheus.remote_write.onprem.receiver`.

Let's see that in action:
```river
local.file "target" {
  filename = "/etc/agent/target"
}

prometheus.scrape "default" {
  targets    = [{ "__address__" = local.file.target.content }]
  forward_to = [prometheus.remote_write.onprem.receiver]
}

prometheus.remote_write "onprem" {
  endpoint {
    url = "http://prometheus:9009/api/prom/push"
  }
}
```

In the previous example, we managed to wire together a very simple pipeline by
writing a few River expressions.

<p align="center">
<img src="../../../../assets/flow_referencing_exports_diagram.svg" alt="Flow of example pipeline" width="500" />
</p>

As with all expressions, once the value is resolved, it must match the [type][]
of the attribute being assigned to. While users can only configure attributes
using the basic River types, the exports of components can also take on special
internal River types such as Secrets or Capsules, which expose different
functionality.


[type]: {{< relref "./types_and_values.md" >}}
