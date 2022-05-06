package flow

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rfratto/gohcl"
)

// File holds the contents of a parsed Flow file.
type File struct {
	Name string    // File name given to ReadFile.
	HCL  *hcl.File // Raw HCL file.

	LogLevel  config.LogLevel
	LogFormat config.LogFormat

	// TODO(rfratto): server block for TLS settings

	// Components holds the list of raw HCL blocks describing components. The
	// Flow controller can interpret this block.
	Components hcl.Blocks
}

// ReadFile parses the HCL file specified by bb into a File. name should be the
// name of the file used for reporting errors.
func ReadFile(name string, bb []byte) (*File, hcl.Diagnostics) {
	file, diags := hclsyntax.ParseConfig(bb, name, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}

	var root rootBlock
	decodeDiags := gohcl.DecodeBody(file.Body, nil, &root)
	diags = diags.Extend(decodeDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	blockSchema := component.RegistrySchema()
	content, remainDiags := root.Remain.Content(blockSchema)
	diags = diags.Extend(remainDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	return &File{
		Name: name,

		LogLevel:   root.LogLevel,
		LogFormat:  root.LogFormat,
		Components: content.Blocks,

		HCL: file,
	}, nil
}

type rootBlock struct {
	LogLevel  config.LogLevel  `hcl:"log_level,optional"`
	LogFormat config.LogFormat `hcl:"log_format,optional"`

	Remain hcl.Body `hcl:",remain"`
}

var defaultRootBlock = rootBlock{
	LogLevel:  config.LogLevelDefault,
	LogFormat: config.LogFormatDefault,
}

var _ gohcl.Decoder = (*rootBlock)(nil)

func (rb *rootBlock) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*rb = defaultRootBlock

	type root rootBlock
	return gohcl.DecodeBody(body, ctx, (*root)(rb))
}
