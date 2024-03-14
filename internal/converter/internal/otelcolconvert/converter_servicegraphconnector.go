package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/connector/servicegraph"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/servicegraphconnector"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/servicegraphprocessor"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, servicegraphConnectorConverter{})
}

type servicegraphConnectorConverter struct{}

func (servicegraphConnectorConverter) Factory() component.Factory {
	return servicegraphconnector.NewFactory()
}

func (servicegraphConnectorConverter) InputComponentName() string {
	return "otelcol.connector.servicegraph"
}

func (servicegraphConnectorConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	// TODO(@tpaschalis) In the version of the OpenTelemetry Collector Contrib
	// we're depending on, the Config still hasn't moved from the
	// servicegraphprocessor to the servicegraphconnector package. Once we
	// update the dependency, we should update the package selector
	// accordingly.
	args := toServicegraphConnector(state, id, cfg.(*servicegraphprocessor.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "connector", "servicegraph"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toServicegraphConnector(state *state, id component.InstanceID, cfg *servicegraphprocessor.Config) *servicegraph.Arguments {
	if cfg == nil {
		return nil
	}
	var (
		nextMetrics = state.Next(id, component.DataTypeMetrics)
	)

	return &servicegraph.Arguments{
		LatencyHistogramBuckets: cfg.LatencyHistogramBuckets,
		Dimensions:              cfg.Dimensions,
		Store: servicegraph.StoreConfig{
			MaxItems: cfg.Store.MaxItems,
			TTL:      cfg.Store.TTL,
		},
		CacheLoop:           cfg.CacheLoop,
		StoreExpirationLoop: cfg.StoreExpirationLoop,
		Output: &otelcol.ConsumerArguments{
			Metrics: toTokenizedConsumers(nextMetrics),
		},
	}
}
