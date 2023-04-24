---
aliases:
- ../../configuration-language/standard-library/coalesce/
title: coalesce
---

# coalesce

`coalesce` takes any number of arguments and returns the first one that isn't null, an empty string, empty list or empty object. It can
be combined with other functions to provide a default.

## Examples

```
> coalesce("a", "b")
a
> coalesce("", "b")
b
> coalesce(env("DOES_NOT_EXIST"), "c")
c
```
