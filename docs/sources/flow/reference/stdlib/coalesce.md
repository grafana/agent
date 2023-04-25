---
aliases:
- ../../configuration-language/standard-library/coalesce/
title: coalesce
---

# coalesce

`coalesce` takes any number of arguments and returns the first one that isn't null, an empty string, empty list, or an empty object.
It is useful for obtaining a default value, such as if an environment variable isn't defined.
If no argument is non-empty or non-zero, the last argument is returned. 

## Examples

```
> coalesce("a", "b")
a
> coalesce("", "b")
b
> coalesce(env("DOES_NOT_EXIST"), "c")
c
```
