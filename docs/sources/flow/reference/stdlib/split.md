---
aliases:
- ../../configuration-language/standard-library/split/
title: split
---

# split

`split` produces a list by dividing a given string at all occurrences of a given separator.

```
split(list, separator)
```

## Examples

```
> split(",", "foo,bar,baz")
["foo", "bar", "baz"]

> split(",", "foo")
["foo"]

> split(",", "")
[""]
```
