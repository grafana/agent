# targets.mutate

The `targets.mutate` component rewrites the label set of the input
targets by applying one or more `relabel_config` steps. If no relabeling
steps are defined, then the input targets will be exported as-is.

The most common use of `targets.mutate` is to filter Prometheus targets
or standardize the label set that will be passed to a downstream component.
The `relabel_config` blocks will be applied to the label set of each target in
order of their appearance in the configuration file.

Multiple `targets.mutate` components can be specified by giving them
different name labels like "keep-backend-only" in the following example.

## Example

```hcl
targets "mutate" "keep-backend-only" {
  targets = [ 
    { "__meta_foo" = "foo", "__address__" = "localhost", "instance" = "one",   "app" = "backend"  },
    { "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "two",   "app" = "database" },
    { "__meta_baz" = "baz", "__address__" = "localhost", "instance" = "three", "app" = "frontend" }
  ]
  
  relabel_config {
    source_labels = ["__address__", "instance"]
    separator     = "/"
    target_label  = "destination"
    action        = "replace"
  } 
  
  relabel_config {
    source_labels = ["app"]
    action        = "keep"
    regex         = "backend"
  }
}
```

## Arguments

The following arguments are supported and can be referenced by other
components:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
targets | list(map(string)) | The targets to mutate. | | **yes**
relabel_config | RelabelConfig | The relabeling steps to apply. | | no


### RelabelConfig
The `relabel_config` block contains the definition of any relabeling rules
that can be applied to an input target.

The following arguments can be used to configure a `relabel_config` block.
All arguments are optional and any omitted fields will take on their default
values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
source_labels | list(string) | The list of labels whose values should be selected. Their content is concatenated using the `separator` and matched against `regex`. | | no
separator     | string       |  The separator used to concatenate the values present in `source_labels`. | ; | no
regex         | string       | A valid RE2 expression with support for parenthesized capture groups. Used to match the extracted value from the combination of the `source_label` and `separator` fields or filter labels during the labelkeep/labeldrop/labelmap actions. | `(.*)` | no
modulus       | uint         | A positive integer used to calculate the modulus of the hashed source label values. | | no
target_label  | string       | Label to which the resulting value will be written to. | | no
replacement   | string       | The value against which a regex replace is performed, if the regex matched the extracted value. Supports previously captured groups. | $1 | no
action        | string       | The relabeling action to perform. | replace | no

Here's a list of the available actions along with a brief description of their usage.

* replace - This action matches `regex` to the concatenated labels. If there's a match, it replaces the content of the `target_label` using the contents of the `replacement` field.
* keep    - This action only keeps the targets where `regex` matches the string extracted using the `source_labels` and `separator`.
* drop    - This action drops the targets where `regex` matches the string extracted using the `source_labels` and `separator`.
* hashmod - This action hashes the concatenated labels, calculates its modulo `modulus` and writes the result to the `target_label`.
* labelmap  - This action matches `regex` against all label names. Any labels that match will be renamed according to the contents of the `replacement` field.
* labeldrop - This action matches `regex` against all label names. Any labels that match will be removed from the target's label set.
* labelkeep - This action matches `regex` against all label names. Any labels that don't match will be removed from the target's label set.

Finally, note that the regex capture groups can be referred to using either the `$1` or `$${1}` notation.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
output | list(map(string)) | The set of targets after applying relabeling.

## Component health

The `targets.mutate` component will only be reported as unhealthy when
given an invalid configuration. In those cases, exported fields will be kept at
their last healthy values.

## Debug information

`targets.mutate` does not expose any component-specific debug information.

### Debug metrics

`targets.mutate` does not expose any component-specific debug metrics.

