---
aliases:
- /docs/agent/latest/flow/configuration-language/expressions/function-calls
title: Function calls
weight: 400
---

# Function Calls
Function calls is one more River feature that lets users build richer
expresisons.

Functions take zero or more arguments as their input and always return a single
value as their output. Functions cannot be constructed by users, but can be
either called from River's standard library, or exported by a component.

## Standard library functions
River contains a [standard library][] of useful functions. Some enable
interaction with the host system (eg. reading from an environment variable), or
allow for more complex expressions (eg. concatenating arrays or decoding JSON
strings into objects).
```river
env("HOME")
json_decode(local.file.cfg.contents)["namespace"]
```

[standard library]: {{< relref "../../reference/stdlib/_index.md" >}}
