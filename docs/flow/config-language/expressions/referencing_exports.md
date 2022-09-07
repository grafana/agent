---
aliases:
- /docs/agent/latest/flow/configuration-language/expressions/referencing-exports
title: Referencing component exports
weight: 200
---

# Referencing component exports
Referencing component exports is what enables River to dynamically configure
and connect components using expressions.

While components can work in isolation, they can be much more versatile when
one component's behavior and data flow is bound to the exports of another,
building a dependency relationship between the two.

## Using references
These references are built by combining the component's name, label and named
export with dots.

For example, the contents of a file exported by the `local.file` component
might be referenced as `local.file.targets.content`, while a
`prometheus.remote_write` component instance might expose a receiver for
metrics like `prometheus.remote_write.onprem.receiver`.

Let's see that in action:
```river
local.file "targets" {
	filename = "/etc/agent/targets" 
}

prometheus.scrape "default" {
	targets    = [{ "__address__" = local.file.target.content }] 
	forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "onprem" {
	remote_write {
		url = "http://prometheus:9009/api/prom/push"
	}
}
```

In the previous example, we managed to wire together a very simple pipeline by
writing a few River expressions.
```
   ┌────────────┐
   │ local.file │
   └──────┬─────┘
          │
          ▼          
┌───────────────────┐
│ prometheus.scrape │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│    prometheus     │
│   remote_write    │
└───────────────────┘
```

As with all expressions, once the value is resolved, it must match the [type][]
of the attribute being assigned to. While users can only configure attributes
using the basic River types, the exports of components can also take on special
internal River types such as Secrets or Capsules, which expose different
functionality.


[type]: {{< relref "./types_and_values.md" >}}
