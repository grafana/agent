---
aliases:
- ../../configuration-language/standard-library/join/
title: join
---

# join

Join all items in an array into a string, using a character as separator.

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
