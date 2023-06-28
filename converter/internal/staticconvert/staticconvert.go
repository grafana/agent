package staticconvert

import (
	"bytes"
	"fmt"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// Convert implements a Prometheus config converter.
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	var staticConfig config.Config
	err := config.LoadBytes(in, false, &staticConfig)
	if err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to parse Static config: %s", err))
		return nil, diags
	}

	if err = staticConfig.Validate(nil); err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to validate Static config: %s", err))
		return nil, diags
	}

	f := builder.NewFile()
	diags = AppendAll(f, &staticConfig)

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags = append(diags, newDiags...)
	return prettyByte, diags
}

type blocks struct {
	discoveryBlocks             []*builder.Block
	discoveryRelabelBlocks      []*builder.Block
	prometheusScrapeBlocks      []*builder.Block
	prometheusRelabelBlocks     []*builder.Block
	prometheusRemoteWriteBlocks []*builder.Block
}

func newBlocks() *blocks {
	return &blocks{
		discoveryBlocks:             []*builder.Block{},
		discoveryRelabelBlocks:      []*builder.Block{},
		prometheusScrapeBlocks:      []*builder.Block{},
		prometheusRelabelBlocks:     []*builder.Block{},
		prometheusRemoteWriteBlocks: []*builder.Block{},
	}
}

// AppendAll analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
// Exports from other components are correctly referenced to build the Flow
// pipeline.
func AppendAll(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	labelCounts := make(map[string]int)
	pb := newBlocks()
	for _, instance := range staticConfig.Metrics.Configs {
		labelCounts[instance.Name]++
		appendPrometheusRemoteWrite(pb, staticConfig.Metrics.Global.RemoteWrite, instance, common.GetUniqueLabel(instance.Name, labelCounts[instance.Name]))
	}

	if staticConfig.Metrics.WALDir != metrics.DefaultConfig.WALDir {
		diags.Add(diag.SeverityLevelWarn, "unsupported config for wal_directory was provided. use the run command flag --storage.path for Flow mode instead.")
	}

	prepareFileBlocks(f, pb)
	return diags
}

// prepareFileBlocks attaches prometheus blocks in a specific order.
//
// Order of blocks:
// 1. Discovery component(s)
// 2. Discovery relabel component(s) (if any)
// 3. Prometheus scrape component(s)
// 4. Prometheus relabel component(s) (if any)
// 5. Prometheus remote_write
func prepareFileBlocks(f *builder.File, pb *blocks) {
	for _, block := range pb.discoveryBlocks {
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.discoveryRelabelBlocks {
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.prometheusScrapeBlocks {
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.prometheusRelabelBlocks {
		f.Body().AppendBlock(block)
	}

	for _, block := range pb.prometheusRemoteWriteBlocks {
		f.Body().AppendBlock(block)
	}
}
