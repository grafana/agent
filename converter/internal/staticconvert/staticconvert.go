package staticconvert

import (
	"bytes"
	"fmt"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/river/token/builder"
	prom_config "github.com/prometheus/prometheus/config"
)

// Convert implements a Static config converter.
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	var staticConfig config.Config
	err := config.LoadBytes(in, false, &staticConfig)
	if err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to parse Static config: %s", err))
		return nil, diags
	}

	if err = staticConfig.Validate(nil); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to validate Static config: %s", err))
		return nil, diags
	}

	f := builder.NewFile()
	diags = AppendAll(f, &staticConfig)

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags = append(diags, newDiags...)
	return prettyByte, diags
}

// AppendAll analyzes the entire static config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
// Exports from other components are correctly referenced to build the Flow
// pipeline.
func AppendAll(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	newDiags := AppendStaticPrometheus(f, staticConfig)
	diags = append(diags, newDiags...)

	// TODO promtail

	// TODO otel

	// TODO other

	return diags
}

func AppendStaticPrometheus(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, instance := range staticConfig.Metrics.Configs {
		promConfig := &prom_config.Config{
			GlobalConfig:       staticConfig.Metrics.Global.Prometheus,
			ScrapeConfigs:      instance.ScrapeConfigs,
			RemoteWriteConfigs: instance.RemoteWrite,
		}

		// There is an edge case unhandled here with label collisions.
		// For example,
		//   metrics config name = "agent_test"
		//   scrape config job_name = "prometheus"
		//
		//   metrics config name = "agent"
		//   scrape config job_name = "test_prometheus"
		//
		//   results in two prometheus.scrape components with the label "agent_test_prometheus"
		newDiags := prometheusconvert.AppendAll(f, promConfig, instance.Name)
		diags = append(diags, newDiags...)
	}

	newDiags := validateMetrics(staticConfig)
	diags = append(diags, newDiags...)
	return diags
}

func validateMetrics(staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	if staticConfig.Metrics.WALDir != metrics.DefaultConfig.WALDir {
		diags.Add(diag.SeverityLevelError, "unsupported config for wal_directory was provided. use the run command flag --storage.path for Flow mode instead.")
	}

	// TODO Static to Flow unsupported features

	return diags
}
