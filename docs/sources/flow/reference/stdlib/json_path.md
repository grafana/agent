---
aliases:
- ../../configuration-language/standard-library/json_path/
title: json_path
---

# json_path

The `json_path` function lookup values using [jsonpath](https://goessner.net/articles/JsonPath/) syntax.

The function expects 2 strings. First string is the json string to lookup values from. Second string is the jsonpath expression.

`json_path` always return a list of values. If the jsonpath expression does not match any values, an empty list is returned.

> Remember to escape double quotes when passing JSON string literals to `json_path`.
>
> For example, the JSON value `{"key": "value"}` is properly represented by the
> string `"{\"key\": \"value\"}"`.

## Examples

```
> json_path("{\"key\": \"value\"}", ".key")
["value"]


> json_path("[{\"name\": \"Department\",\"value\": \"IT\"},{\"name\":\"TestStatus\",\"value\":\"Pending\"}]", "[?(@.name == \"Department\")].value")
["IT"]

> json_path("{\"key\": \"value\"}", ".nonexists")
[]

> json_path("{\"key\": \"value\"}", ".key")[0]
value

```
