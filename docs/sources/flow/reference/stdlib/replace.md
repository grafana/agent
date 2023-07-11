---
aliases:
- ../../configuration-language/standard-library/replace/
title: replace
---

# replace

`replace` searches a string for a substring, and replaces each occurrence of the substring with a replacement string.

```river
replace(string, substring, replacement)
```

## Examples

```river
> replace("1 + 2 + 3", "+", "-")
"1 - 2 - 3"
```
