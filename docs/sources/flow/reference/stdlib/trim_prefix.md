---
aliases:
- ../../configuration-language/standard-library/trim_prefix/
canonical: https://grafana.com/docs/grafana/agent/latest/flow/reference/stdlib/trim_prefix/
title: trim_prefix
---

# trim_prefix

`trim_prefix` removes the prefix from the start of a string. If the string does not start with the prefix, the string is returned unchanged.

## Examples

```river
> trim_prefix("helloworld", "hello")
"world"
```
