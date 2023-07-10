---
aliases:
- ../../configuration-language/standard-library/trim_prefix/
title: trim_prefix
---

# trim_prefix

`trim_prefix` removes the specified prefix from the start of the given string. If the string does not start with the prefix, the string is returned unchanged.

## Examples

```
> trim_prefix("helloworld", "hello")
"world"
```
