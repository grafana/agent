---
aliases:
- ../../configuration-language/standard-library/trim/
title: trim
---

# trim

`trim` removes the specified set of characters from the start and end of the given string.

```
trim(string, str_character_set)
```

Every occurrence of a character in the second argument is removed from the start and end of the string specified in the first argument.

## Examples

```
> trim("?!hello?!", "!?")
"hello"

> trim("foobar", "far")
"oob"

> trim("   hello! world.!  ", "! ")
"hello! world."
```
