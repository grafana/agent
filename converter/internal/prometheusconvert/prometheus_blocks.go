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
	exporterBlocks              []prometheusBlock
	discoveryBlocks             []prometheusBlock
	discoveryRelabelBlocks      []prometheusBlock
	prometheusScrapeBlocks      []prometheusBlock
	prometheusRelabelBlocks     []prometheusBlock
	prometheusRemoteWriteBlocks []prometheusBlock
}

func newPrometheusBlocks() *prometheusBlocks {
	return &prometheusBlocks{
		exporterBlocks:              []prometheusBlock{},
		discoveryBlocks:             []prometheusBlock{},
		discoveryRelabelBlocks:      []prometheusBlock{},
		prometheusScrapeBlocks:      []prometheusBlock{},
		prometheusRelabelBlocks:     []prometheusBlock{},
		prometheusRemoteWriteBlocks: []prometheusBlock{},
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
func (pb *prometheusBlocks) appendToFile(f *builder.File) {
	for _, promBlock := range pb.exporterBlocks {
		f.Body().AppendBlock(promBlock.block)
	}

	for _, promBlock := range pb.discoveryBlocks {
		f.Body().AppendBlock(promBlock.block)
	}

	for _, promBlock := range pb.discoveryRelabelBlocks {
		f.Body().AppendBlock(promBlock.block)
	}

	for _, promBlock := range pb.prometheusScrapeBlocks {
		f.Body().AppendBlock(promBlock.block)
	}

	for _, promBlock := range pb.prometheusRelabelBlocks {
		f.Body().AppendBlock(promBlock.block)
	}

	for _, promBlock := range pb.prometheusRemoteWriteBlocks {
		f.Body().AppendBlock(promBlock.block)
	}
}

func (pb *prometheusBlocks) getScrapeInfo() diag.Diagnostics {
	var diags diag.Diagnostics

	for _, promScrapeBlock := range pb.prometheusScrapeBlocks {
		detail := promScrapeBlock.detail

		for _, promDiscoveryBlock := range pb.discoveryBlocks {
			if strings.HasPrefix(promDiscoveryBlock.label, promScrapeBlock.label) {
				detail = fmt.Sprintln(detail) + fmt.Sprintf("	A %s.%s component", strings.Join(promDiscoveryBlock.name, "."), promDiscoveryBlock.label)
			}
		}

		for _, promDiscoveryRelabelBlock := range pb.discoveryRelabelBlocks {
			if strings.HasPrefix(promDiscoveryRelabelBlock.label, promScrapeBlock.label) {
				detail = fmt.Sprintln(detail) + fmt.Sprintf("	A %s.%s component", strings.Join(promDiscoveryRelabelBlock.name, "."), promDiscoveryRelabelBlock.label)
			}
		}

		for _, promRelabelBlock := range pb.prometheusRelabelBlocks {
			if strings.HasPrefix(promRelabelBlock.label, promScrapeBlock.label) {
				detail = fmt.Sprintln(detail) + fmt.Sprintf("	A %s.%s component", strings.Join(promRelabelBlock.name, "."), promRelabelBlock.label)
			}
		}

		diags.AddWithDetail(diag.SeverityLevelInfo, promScrapeBlock.summary, detail)
	}

	for _, promRemoteWriteBlock := range pb.prometheusRemoteWriteBlocks {
		diags.AddWithDetail(diag.SeverityLevelInfo, promRemoteWriteBlock.summary, promRemoteWriteBlock.detail)
	}

	return diags
}

type prometheusBlock struct {
	block   *builder.Block
	name    []string
	label   string
	summary string
	detail  string
}

func newPrometheusBlock(block *builder.Block, name []string, label string, summary string, detail string) prometheusBlock {
	return prometheusBlock{
		block:   block,
		name:    name,
		label:   label,
		summary: summary,
		detail:  detail,
	}
}
