package build

import (
	"fmt"
	"reflect"

	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/otelcolconvert"
	"github.com/grafana/agent/internal/static/traces"
	otel_component "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"
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

		removeReceiver(otelCfg, "traces", "push_receiver")
		b.translateAutomaticLogging(otelCfg)

		// Let's only prefix things if we are doing more than 1 trace config
		labelPrefix := ""
		if len(b.cfg.Traces.Configs) > 1 {
			labelPrefix = cfg.Name
		}
		b.diags.AddAll(otelcolconvert.AppendConfig(b.f, otelCfg, labelPrefix))
	}
}

func (b *ConfigBuilder) translateAutomaticLogging(otelCfg *otelcol.Config) {
	if _, ok := otelCfg.Processors[otel_component.NewID("automatic_logging")]; !ok {
		return
	}

	removeProcessor(otelCfg, "traces", "automatic_logging")
}

// removeReceiver removes a receiver from the otel config. The
// push_receiver is an implementation detail for static traces that is not
// necessary for the flow configuration.
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

// removeProcessor removes a processor from the otel config.
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
