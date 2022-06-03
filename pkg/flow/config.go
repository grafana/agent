package flow

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rfratto/gohcl"
)

// File holds the contents of a parsed Flow file.
type File struct {
	Name string    // File name given to ReadFile.
	HCL  *hcl.File // Raw HCL file.

	Logging logging.Options

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

	if root.Logger == nil {
		defaults := logging.DefaultOptions
		root.Logger = &defaults
	}

	return &File{
		Name:       name,
		HCL:        file,
		Logging:    *root.Logger,
		Components: content.Blocks,
	}, nil
}

type rootBlock struct {
	Logger *logging.Options `hcl:"logging,block"`

	// TODO(rfratto): server block for TLS settings

	Remain hcl.Body `hcl:",remain"`
}

var defaultRootBlock = rootBlock{}

var _ gohcl.Decoder = (*rootBlock)(nil)

func (rb *rootBlock) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*rb = defaultRootBlock

	type root rootBlock
	return gohcl.DecodeBody(body, ctx, (*root)(rb))
}
