# discovery.transformer

The `discovery.transformer` component rewrites the label set of the input
targets by applying one or more [`relabel_config`][] steps. If no relabeling
steps are defined, then the input targets will be exported as-is.

The most common use of `discovery.transformer` is to filter Prometheus targets
or standardize the label set that will be passed to a downstream component.
The `relabel_config` blocks will be applied to the label set of each target in
order of their appearance in the configuration file.

Multiple `discovery.transformer` components can be specified by giving them
different name labels like "keep-backend-only" in the following example.

## Example

```hcl
discovery "transformer" "keep-backend-only" {
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
    action = "keep"
    regex  = "backend"
  }
}
```

## Arguments

The following arguments are supported and can be referenced by other
components:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets` | `[]map[string]string` | The input targets | | no
`relabel_config` | `RelabelConfig` | The relabeling steps to apply | | no


### Targets
The `targets` field contains the array of input targets. Each target consists
of one or more string key-value pairs, referred to as its label set.

### RelabelConfig
The `relabel_config` block contains the definition of any relabeling rules
that can be applied to an input target. If you want an overview of how
relabeling works you can refer to the [Prometheus docs][] or this [blogpost][].

#### Argument Reference
The following arguments can be used to configure a `relabel_config` block.
All arguments are optional and ny omitted fields will take on their default
values.

* `source_labels` - The list of labels whose values should be selected. Their content is concatenated using the `separator` and matched against `regex`.
* `separator` - Used to concatenate the values selected from `source_labels`. Defaults to `;`.
* `regex` - A valid RE2 expression with support for parenthesized capture groups. Used to match the extracted value from the combination of the `source_label` and `separator` fields or filter labels during the labelkeep/labeldrop/labelmap actions. Defaults to `(.*)`.
* `modulus` - A positive integer used to calculate the modulus of the hashed source label values.
* `target_label` - Label to which the resulting value will be written to.
* `replacement` - The value against which a regex replace is performed, if the regex matched the extracted value. Supports previously captured groups. Defaults to `$1`.
* `action` - The relabeling action to perform. Defaults to "replace".

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`output_targets` | `[]map[string]string` | The resulting targets' label set.

The number of targets exposed in `output_targets` will be less or equal than
the number of input targets, as some of them may be dropped by the
relabeling process.

## Component health

Any `discovery.transformer` component will be reported as healthy whenever
it is able to correctly apply all relabeling steps to the input targets.
If the either the input `targets`, or `relabel_config` blocks could not
be parsed or the relabeling steps could not be applied, the component will
be reported as unhealthy. In those cases, exported fields will be kept at
the last healthy values.

## Debug information

`discovery.transformer` does not expose any component-specific debug information.

### Debug metrics

`discovery.transformer` does not expose any component-specific debug metrics.

[`relabel_config`]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
[Prometheus docs]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
[blogpost]: https://grafana.com/blog/2022/03/21/how-relabeling-in-prometheus-works/
