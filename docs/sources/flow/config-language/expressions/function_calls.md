---
aliases:
- ../../configuration-language/expressions/function-calls/
canonical: https://grafana.com/docs/agent/latest/flow/config-language/expressions/function_calls/
title: Function calls
weight: 400
---

# Function calls
Function calls is one more River feature that lets users build richer
expressions.

Functions take zero or more arguments as their input and always return a single
value as their output. Functions cannot be constructed by users, but can be
either called from River's standard library, or when exported by a component.

In case a function fails, the expression will not be evaluated and an error
will be reported.

## Standard library functions
River contains a [standard library][] of useful functions. Some enable
interaction with the host system (e.g. reading from an environment variable), or
allow for more complex expressions (e.g. concatenating arrays or decoding JSON
strings into objects).
```river
env("HOME")
json_decode(local.file.cfg.contents)["namespace"]
```

[standard library]: {{< relref "../../reference/stdlib/_index.md" >}}
