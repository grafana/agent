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
`relabel_config` | `RelabelConfig` | The relabeling steps to apply | `DefaultRelabelConfig` | no


### Targets
The `targets` field contains the array of input targets. Each target consists
of one or more string key-value pairs, referred to as its label set.

### RelabelConfig
The `relabel_config` block contains the definition of any relabeling rules
that can be applied to an input target. If you want an overview of how
relabeling works you can refer to the [Prometheus docs][] or this [blogpost][].

```go
type RelabelConfig struct {
	SourceLabels []string `hcl:"source_labels,optional"`
	Separator    string   `hcl:"separator,optional"`
	Regex        Regexp   `hcl:"regex,optional"`
	Modulus      uint64   `hcl:"modulus,optional"`
	TargetLabel  string   `hcl:"target_label,optional"`
	Replacement  string   `hcl:"replacement,optional"`
	Action       Action   `hcl:"action,optional"`
}
```

### DefaultRelabelConfig
The `DefaultRelabelConfig` object contains the default values for omitted fields
when decoding an HCL block to a RelabelConfig struct.

```go
var DefaultRelabelConfig = RelabelConfig{
	Action:      Replace,
	Separator:   ";",
	Regex:       "(.*)",
	Replacement: "$1",
}
```

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
