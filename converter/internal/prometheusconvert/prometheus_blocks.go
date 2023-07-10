package prometheusconvert

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// prometheusBlocks is a type for categorizing River Blocks before appending
// them to a River File. This gives control over the order they are written
// versus appending them in the order the Blocks are created.
type prometheusBlocks struct {
	discoveryBlocks             []*builder.Block
	discoveryRelabelBlocks      []*builder.Block
	prometheusScrapeBlocks      []*builder.Block
	prometheusRelabelBlocks     []*builder.Block
	prometheusRemoteWriteBlocks []*builder.Block
}

func newPrometheusBlocks() *prometheusBlocks {
	return &prometheusBlocks{
		discoveryBlocks:             []*builder.Block{},
		discoveryRelabelBlocks:      []*builder.Block{},
		prometheusScrapeBlocks:      []*builder.Block{},
		prometheusRelabelBlocks:     []*builder.Block{},
		prometheusRemoteWriteBlocks: []*builder.Block{},
	}
}

// appendToFile attaches prometheus blocks in a specific order.
//
// Order of blocks:
// 1. Discovery component(s)
// 2. Discovery relabel component(s) (if any)
// 3. Prometheus scrape component(s)
// 4. Prometheus relabel component(s) (if any)
// 5. Prometheus remote_write
func (pb *prometheusBlocks) appendToFile(f *builder.File) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, block := range pb.discoveryBlocks {
		diags.Add(diag.SeverityLevelInfo, fmt.Sprintf("Appending %s.%s", strings.Join(block.Name, "."), block.Label))
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.discoveryRelabelBlocks {
		diags.Add(diag.SeverityLevelInfo, fmt.Sprintf("Appending %s.%s", strings.Join(block.Name, "."), block.Label))
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.prometheusScrapeBlocks {
		diags.Add(diag.SeverityLevelInfo, fmt.Sprintf("Appending %s.%s", strings.Join(block.Name, "."), block.Label))
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.prometheusRelabelBlocks {
		diags.Add(diag.SeverityLevelInfo, fmt.Sprintf("Appending %s.%s", strings.Join(block.Name, "."), block.Label))
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.prometheusRemoteWriteBlocks {
		diags.Add(diag.SeverityLevelInfo, fmt.Sprintf("Appending %s.%s", strings.Join(block.Name, "."), block.Label))
		f.Body().AppendBlock(block)
	}

	return diags
}
