package staticconvert

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/converter/internal/promtailconvert"
	"github.com/grafana/agent/converter/internal/staticconvert/internal/build"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/logs"
	promtail_config "github.com/grafana/loki/clients/pkg/promtail/config"
	"github.com/grafana/loki/clients/pkg/promtail/limit"
	"github.com/grafana/loki/clients/pkg/promtail/targets/file"
	"github.com/grafana/river/scanner"
	"github.com/grafana/river/token/builder"
	prom_config "github.com/prometheus/prometheus/config"

	_ "github.com/grafana/agent/pkg/integrations/install" // Install integrations
)

// Convert implements a Static config converter.
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	staticConfig, err := config.LoadFromFunc(fs, []string{"-config.file", "convert", "-config.expand-env"}, func(_, _ string, expandEnvVars bool, c *config.Config) error {
		return config.LoadBytes(in, expandEnvVars, c)
	})

	if err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to parse Static config: %s", err))
		return nil, diags
	}

	f := builder.NewFile()
	diags = AppendAll(f, staticConfig)
	diags.AddAll(common.ValidateNodes(f))

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags.AddAll(newDiags)
	return prettyByte, diags
}

// AppendAll analyzes the entire static config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
// Exports from other components are correctly referenced to build the Flow
// pipeline.
func AppendAll(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddAll(appendStaticPrometheus(f, staticConfig))
	diags.AddAll(appendStaticPromtail(f, staticConfig))
	diags.AddAll(appendStaticIntegrationsV1(f, staticConfig))
	// TODO integrations v2
	// TODO otel

	diags.AddAll(validate(staticConfig))

	return diags
}

func appendStaticPrometheus(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, instance := range staticConfig.Metrics.Configs {
		promConfig := &prom_config.Config{
			GlobalConfig:       staticConfig.Metrics.Global.Prometheus,
			ScrapeConfigs:      instance.ScrapeConfigs,
			RemoteWriteConfigs: instance.RemoteWrite,
		}

		jobNameToCompLabelsFunc := func(jobName string) string {
			name := fmt.Sprintf("metrics_%s", instance.Name)
			if jobName != "" {
				name += fmt.Sprintf("_%s", jobName)
			}

			name, err := scanner.SanitizeIdentifier(name)
			if err != nil {
				diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to sanitize job name: %s", err))
			}

			return name
		}

		// There is an edge case here with label collisions that will be caught
		// by a validation [common.ValidateNodes].
		// For example,
		//   metrics config name = "agent_test"
		//   scrape config job_name = "prometheus"
		//
		//   metrics config name = "agent"
		//   scrape config job_name = "test_prometheus"
		//
		//   results in two prometheus.scrape components with the label "metrics_agent_test_prometheus"
		diags.AddAll(prometheusconvert.AppendAllNested(f, promConfig, jobNameToCompLabelsFunc, []discovery.Target{}, nil))
	}

	return diags
}

func appendStaticPromtail(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	if staticConfig.Logs == nil {
		return diags
	}

	for _, logConfig := range staticConfig.Logs.Configs {
		promtailConfig := logs.DefaultConfig()
		promtailConfig.Global = promtail_config.GlobalConfig{FileWatch: staticConfig.Logs.Global.FileWatch}
		promtailConfig.ClientConfigs = logConfig.ClientConfigs
		promtailConfig.PositionsConfig = logConfig.PositionsConfig
		promtailConfig.ScrapeConfig = logConfig.ScrapeConfig
		promtailConfig.TargetConfig = logConfig.TargetConfig
		promtailConfig.LimitsConfig = logConfig.LimitsConfig

		// We need to set this when empty so the promtail converter doesn't think it has been overridden
		if promtailConfig.Global == (promtail_config.GlobalConfig{}) {
			promtailConfig.Global.FileWatch = file.DefaultWatchConig
		}

		if promtailConfig.LimitsConfig == (limit.Config{}) {
			promtailConfig.LimitsConfig = promtailconvert.DefaultLimitsConfig()
		}

		// There is an edge case here with label collisions that will be caught
		// by a validation [common.ValidateNodes].
		// For example,
		//   logs config name = "agent_test"
		//   scrape config job_name = "promtail"
		//
		//   logs config name = "agent"
		//   scrape config job_name = "test_promtail"
		//
		//   results in two prometheus.scrape components with the label "logs_agent_test_promtail"
		diags = promtailconvert.AppendAll(f, &promtailConfig, "logs_"+logConfig.Name, diags)
	}

	return diags
}

func appendStaticIntegrationsV1(f *builder.File, staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	b := build.NewIntegrationsV1ConfigBuilder(f, &diags, staticConfig, &build.GlobalContext{LabelPrefix: "integrations"})
	b.Build()

	return diags
}
