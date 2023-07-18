---
aliases:
- ../../configuration-language/standard-library/join/
canonical: https://grafana.com/docs/grafana/agent/latest/flow/reference/stdlib/join/
title: join
---

# join

`join` all items in an array into a string, using a character as separator.

```river
join(list, separator)
```

## Examples

```river
> join(["foo", "bar", "baz"], "-")
"foo-bar-baz"
> join(["foo", "bar", "baz"], ", ")
"foo, bar, baz"
> join(["foo"], ", ")
"foo"
```
