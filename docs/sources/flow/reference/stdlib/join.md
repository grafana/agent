---
aliases:
- ../../configuration-language/standard-library/join/
title: join
---

# join

`join` produces a string by concatenating all of the elements of the specified list of strings with the specified separator.

```
join(list, separator)
```

## Examples

```
> join(["foo", "bar", "baz"], "-")
"foo-bar-baz"
> join(["foo", "bar", "baz"], ", ")
"foo, bar, baz"
> join(["foo"], ", ")
"foo"
```
