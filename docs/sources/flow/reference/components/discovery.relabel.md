---
title: discovery.relabel
---

# discovery.relabel

In Flow, targets are defined as sets of key-value pairs called _labels_.

`discovery.relabel` rewrites the label set of the input targets by applying one
or more relabeling rules. If no rules are defined, then the input targets are
exported as-is.

The most common use of `discovery.relabel` is to filter targets or standardize
the target label set that is passed to a downstream component. The `rule`
blocks are applied to the label set of each target in order of their appearance
in the configuration file.

Target labels which start with a double underscore `__` are considered
internal, and may be removed by other Flow components prior to telemetry
collection. To retain any of these labels, use a `labelmap` action to remove
the prefix, or remap them to a different name. Service discovery mechanisms
usually group their labels under `__meta_*`. For example, the
discovery.kubernetes component populates a set of `__meta_kubernetes_*` labels
to provide information about the discovered Kubernetes resources. If a
relabeling rule needs to store a label value temporarily, for example as the
input to a subsequent step, use the `__tmp` label name prefix, as it is
guaranteed to never be used.

Multiple `discovery.relabel` components can be specified by giving them
different labels.

## Usage

```
discovery.relabel "LABEL" {
  targets = TARGET_LIST

  rule {
    ...
  }

  ...
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets` | `list(map(string))` | Targets to relabel | | **yes**

## Blocks

The following blocks are supported inside the definition of
`discovery.relabel`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
rule | [rule][] | Relabeling rules to apply to targets. | no

[rule]: #rule-block

### rule block

The `rule` block contains the definition of any relabeling rules that
can be applied to an input target. If more than one `rule` block is
defined within `discovery.relabel`, the transformations are applied
in top-down order.

The following arguments can be used to configure a `rule` block.
All arguments are optional and any omitted fields take on their default
values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`source_labels` | `list(string)` | The list of labels whose values should be selected. Their content is concatenated using the `separator` and matched against `regex`. | | no
`separator`     | `string`       |  The separator used to concatenate the values present in `source_labels`. | `;` | no
`regex`         | `string`       | A valid RE2 expression with support for parenthesized capture groups. Used to match the extracted value from the combination of the `source_label` and `separator` fields or filter labels during the labelkeep/labeldrop/labelmap actions. | `(.*)` | no
`modulus`       | `uint`         | A positive integer used to calculate the modulus of the hashed source label values. | | no
`target_label`  | `string`       | Label to which the resulting value are written to. | | no
`replacement`   | `string`       | The value against which a regex replace is performed, if the regex matched the extracted value. Supports previously captured groups. | `$1` | no
`action`        | `string`       | The relabeling action to perform. | `replace` | no

Here's a list of the available actions along with a brief description of their usage.

* `replace` - This action matches `regex` to the concatenated labels. If there's a match, it replaces the content of the `target_label` using the contents of the `replacement` field.
* `keep`    - This action only keeps the targets where `regex` matches the string extracted using the `source_labels` and `separator`.
* `drop`    - This action drops the targets where `regex` matches the string extracted using the `source_labels` and `separator`.
* `hashmod` - This action hashes the concatenated labels, calculates its modulo `modulus` and writes the result to the `target_label`.
* `labelmap`  - This action matches `regex` against all label names. Any labels that match are renamed according to the contents of the `replacement` field.
* `labeldrop` - This action matches `regex` against all label names. Any labels that match are removed from the target's label set.
* `labelkeep` - This action matches `regex` against all label names. Any labels that don't match are removed from the target's label set.

Finally, note that the regex capture groups can be referred to using either the `$1` or `$${1}` notation.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`output` | `list(map(string))` | The set of targets after applying relabeling.

## Component health

`discovery.relabel` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.relabel` does not expose any component-specific debug information.

### Debug metrics

`discovery.relabel` does not expose any component-specific debug metrics.

## Example

```river
discovery.relabel "keep_backend_only" {
  targets = [
    { "__meta_foo" = "foo", "__address__" = "localhost", "instance" = "one",   "app" = "backend"  },
    { "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "two",   "app" = "database" },
    { "__meta_baz" = "baz", "__address__" = "localhost", "instance" = "three", "app" = "frontend" },
  ]

  rule {
    source_labels = ["__address__", "instance"]
    separator     = "/"
    target_label  = "destination"
    action        = "replace"
  }

  rule {
    source_labels = ["app"]
    action        = "keep"
    regex         = "backend"
  }
}
```


