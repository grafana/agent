---
aliases:
- ../../configuration-language/standard-library/regex_quote/
title: regex_quote
---

# regex_quote

The `regex_quote` function returns a string that escapes all regular expression metacharacters inside the argument 
text. The returned string is a regular expression matching the literal text.

## Examples

```
> regex_escape("ip-10-0-69-8.ec2.internal")
"ip-10-0-69-8\.ec2\.internal"
```
