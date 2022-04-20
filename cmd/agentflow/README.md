# Agent Flow Expressions

You can run Agent Flow from the root of the repository with:

```
go run ./cmd/agentflow -config.file ./cmd/agentflow/example-config.flow
```

See [the example config](./example-config.flow) for the contents of the config file referenced above.

## Adding a new component

The [`component` package](../../component/component.go) describes interfaces
which components may implement. [`metrics_scraper`][] serves as a good example
component.

Components must exist somewhere in the Go import path to be usable, just like
integrations. Import your new package in
[`pkg/flow/install`](../../pkg/flow/install/) to guarantee it gets imported.

## Decoding tricks

This branch uses a [gohcl fork](https://pkg.go.dev/github.com/rfratto/gohcl)
which includes support for custom decoding. Types that implement
[`encoding.TextMarshaler`](https://pkg.go.dev/encoding#TextMarshaler) and
[`encoding.TextUnmarshaler`](https://pkg.go.dev/encoding#TextUnmarshaler) will
be converted to and from strings during HCL conversion.

HCL blocks may also implement
[`gohcl.Decoder`](https://pkg.go.dev/github.com/rfratto/gohcl#Decoder) to
implement defaulting and config validation. The [`metrics_scraper`][] config
struct does this.

### HCL tags

HCL tags are used to map to HCL attributes (i.e., settings) or blocks (i.e.,
objects):

* Required attribute: `hcl:"NAME"`
* Optional attribute: `hcl:"NAME,optional"`
* Block: `hcl:"NAME,block"`

If the field of a block type is a slice, that block may be defined more than
once. See [`metrics_forwarder`][] for an example of using block types with
remote_write.

[`metrics_scraper`]: ../../component/metrics-scraper/scraper.go
[`metrics_forwarder`]: ../../component/metrics-forwarder/forwarder.go
