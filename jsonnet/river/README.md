# Jsonnet `river` library

The Jsonnet `river` library makes it possible to return River-formatted config
files using Jsonnet.

To manifest a River configuration file, call `river.manifestRiver(value)`.

Field names from objects are expected to follow one of the three forms:

* `<name>` for River attributes (e.g., `foobar`).
* `block <name>` for unlabeled River blocks (e.g., `block exporter.node`)
* `block <name> <label>` for labeled River blocks (.e.g, `block metrics.remote_write default`).

Instead of following these naming conventions, helper functions are provided to
make it easier:

* `river.attr(name)` returns a field name that can be used as an attribute.
* `river.block(name, label="")` returns a field name that represents a block.

In addition to the helper functions, `river.expr(literal)` is used to inject a
literal River expression, so that `river.expr('env("HOME")')` is manifested as
the literal River expression `env("HOME")`.

## Example

```jsonnet
river.manifestRiver({
  attr_1: "Hello, world!",

  [river.block("some_block", "foobar")]: {
    expr: river.expr('env("HOME")'),
    inner_attr_1: [0, 1, 2, 3],
    inner_attr_2: {
      first_name: "John",
      last_name: "Smith",
    },
  },
})
```

results in

```river
attr_1 = "Hello, world"
some_block "foobar" {
  expr = env("HOME")
  inner_attr_1 = [0, 1, 2, 3]
  inner_attr_2 = {
    "first_name" = "John",
    "last_name" = "Smith",
  }
}
```
