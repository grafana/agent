---
aliases:
- /docs/agent/latest/concepts/configuration-language
title: Configuration language
weight: 300
---

# Configuration language

The _configuration language_ refers to the language used in configuration files
which define and configure components to run.

The configuration language is called River, a Terraform/HCL-inspired language:

```river
metrics.scrape "default" {
  targets = [{
    "__address__" = "demo.robustperception.io:9090",
  }]
  forward_to = [metrics.remote_write.default.receiver]
}

metrics.remote_write "default" {
  remote_write {
    url = "http://localhost:9009/api/prom/push"
  }
}
```

River was designed with two requirements in mind:

* _Fast_: The configuration language must be fast so the component controller
  can evaluate changes as quickly as possible.
* _Simple_: The configuration language must be easy to read and write to
  minimize the learning curve.
* _Debuggable_: The configuration language must give detailed information when
  there's a mistake in the config file.

Our dedicated [Configuration language][config-docs] section documents River in
detail.

[config-docs]: {{< relref "../config-language/" >}}
