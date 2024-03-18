package build

import (
	"fmt"
	"reflect"

	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/otelcolconvert"
	"github.com/grafana/agent/internal/converter/internal/prometheusconvert"
	"github.com/grafana/agent/internal/static/traces"
	"github.com/grafana/river/scanner"
	prom_config "github.com/prometheus/prometheus/config"
	otel_component "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/otelcol"
	"gopkg.in/yaml.v3"
)

func (b *ConfigBuilder) appendTraces() {
	if reflect.DeepEqual(b.cfg.Traces, traces.Config{}) {
		return
	}

	for _, cfg := range b.cfg.Traces.Configs {
		otelCfg, err := cfg.OtelConfig()
		if err != nil {
			b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to load otelConfig from agent traces config: %s", err))
			continue
		}

		// Only prefix component labels if we are doing more than 1 trace config.
		labelPrefix := ""
		if len(b.cfg.Traces.Configs) > 1 {
			labelPrefix = cfg.Name
		}

		// Remove the push receiver which is an implementation detail for static mode and unnecessary for the otel config.
		removeReceiver(otelCfg, "traces", "push_receiver")

		b.translateAutomaticLogging(otelCfg, cfg)
		b.translatePromSDProcessor(otelCfg, cfg, labelPrefix)

		b.diags.AddAll(otelcolconvert.AppendConfig(b.f, otelCfg, labelPrefix))
	}
}

func (b *ConfigBuilder) translateAutomaticLogging(otelCfg *otelcol.Config, cfg traces.InstanceConfig) {
	if _, ok := otelCfg.Processors[otel_component.NewID("automatic_logging")]; !ok {
		return
	}

	if cfg.AutomaticLogging.Backend == "stdout" {
		b.diags.Add(diag.SeverityLevelWarn, "automatic_logging for traces has no direct flow equivalent. "+
			"A best effort translation has been made to otelcol.exporter.logging but the behavior will differ.")
	} else {
		b.diags.Add(diag.SeverityLevelError, "automatic_logging for traces has no direct flow equivalent. "+
			"A best effort translation can be made which only outputs to stdout and not directly to loki by bypassing errors.")
	}

	// Add the logging exporter to the otel config with default values
	otelCfg.Exporters[otel_component.NewID("logging")] = loggingexporter.NewFactory().CreateDefaultConfig()

	// Add the logging exporter to all pipelines
	for _, pipeline := range otelCfg.Service.Pipelines {
		pipeline.Exporters = append(pipeline.Exporters, otel_component.NewID("logging"))
	}

	// Remove the custom automatic_logging processor
	removeProcessor(otelCfg, "traces", "automatic_logging")
}

func (b *ConfigBuilder) translatePromSDProcessor(otelCfg *otelcol.Config, cfg traces.InstanceConfig, labelPrefix string) {
	if _, ok := otelCfg.Processors[otel_component.NewID("prom_sd_processor")]; !ok {
		return
	}

	out, err := yaml.Marshal(cfg.ScrapeConfigs)
	if err != nil {
		b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("unable to marshal scrapeConfigs interface{} to yaml: %s", err))
		return
	}

	scrapeConfigs := make([]*prom_config.ScrapeConfig, 0)
	err = yaml.Unmarshal(out, &scrapeConfigs)
	if err != nil {
		b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("unable to unmarshal bytes to []*config.ScrapeConfig: %s", err))
	}

	promConfig := &prom_config.Config{
		ScrapeConfigs: scrapeConfigs,
	}

	// Remove the prom_sd_processor processor which is an implementation detail for static mode and unnecessary for the otel config.
	removeProcessor(otelCfg, "traces", "prom_sd_processor")

	jobNameToCompLabelsFunc := func(jobName string) string {
		name := labelPrefix
		if jobName != "" {
			name += fmt.Sprintf("_%s", jobName)
		}

		if name == "" {
			name = "default"
		}

		name, err := scanner.SanitizeIdentifier(name)
		if err != nil {
			b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to sanitize job name: %s", err))
		}

		return name
	}

	prometheusconvert.AppendAllNested(b.f, promConfig, jobNameToCompLabelsFunc, nil, nil)

	// // Add the prom_sd_processor processor to the otel config with default values
	// otelCfg.Processors[otel_component.NewID("prom_sd_processor")] = traces.NewPromSDProcessorFactory().CreateDefaultConfig()

	// // Add the prom_sd_processor processor to all pipelines
	// for _, pipeline := range otelCfg.Service.Pipelines {
	// 	pipeline.Processors = append(pipeline.Processors, otel_component.NewID("prom_sd_processor"))
	// }
}

// removeReceiver removes a receiver from the otel config for a specific pipeline type.
func removeReceiver(otelCfg *otelcol.Config, pipelineType otel_component.Type, receiverType otel_component.Type) {
	if _, ok := otelCfg.Receivers[otel_component.NewID(receiverType)]; !ok {
		return
	}

	delete(otelCfg.Receivers, otel_component.NewID(receiverType))
	spr := make([]otel_component.ID, 0, len(otelCfg.Service.Pipelines[otel_component.NewID(pipelineType)].Receivers)-1)
	for _, r := range otelCfg.Service.Pipelines[otel_component.NewID(pipelineType)].Receivers {
		if r != otel_component.NewID(receiverType) {
			spr = append(spr, r)
		}
	}
	otelCfg.Service.Pipelines[otel_component.NewID(pipelineType)].Receivers = spr
}

// removeProcessor removes a processor from the otel config for a specific pipeline type.
func removeProcessor(otelCfg *otelcol.Config, pipelineType otel_component.Type, processorType otel_component.Type) {
	if _, ok := otelCfg.Processors[otel_component.NewID(processorType)]; !ok {
		return
	}

	delete(otelCfg.Processors, otel_component.NewID(processorType))
	spr := make([]otel_component.ID, 0, len(otelCfg.Service.Pipelines[otel_component.NewID(pipelineType)].Processors)-1)
	for _, r := range otelCfg.Service.Pipelines[otel_component.NewID(pipelineType)].Processors {
		if r != otel_component.NewID(processorType) {
			spr = append(spr, r)
		}
	}
	otelCfg.Service.Pipelines[otel_component.NewID(pipelineType)].Processors = spr
}
