package spanhostmetrics

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// --- Config

type Config struct {
	AttributeNames []string `mapstructure:"attribute_names"`
}

// --- Factory

const (
	// this is the name used to refer to the connector in the config.yaml
	typeStr = "hostspanmetrics"
)

// NewFactory creates a factory for example connector.
func NewFactory() connector.Factory {
	// OpenTelemetry connector factory to make a factory for connectors

	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetricsConnector, component.StabilityLevelAlpha))
}

func createDefaultConfig() component.Config {
	return &Config{
		AttributeNames: []string{"k8s.node.name", "host.name"},
	}
}

// createTracesToMetricsConnector defines the consumer type of the connector
// We want to consume traces and export metrics, therefore, define nextConsumer as metrics, since consumer is the next component in the pipeline
func createTracesToMetricsConnector(ctx context.Context, params connector.CreateSettings, cfg component.Config, nextConsumer consumer.Metrics) (connector.Traces, error) {
	c, err := newConnector(params.Logger, cfg)
	if err != nil {
		return nil, err
	}
	c.metricsConsumer = nextConsumer
	return c, nil
}

// --- Connector

// schema for connector
type connectorImp struct {
	config          Config
	metricsConsumer consumer.Metrics
	logger          *zap.Logger
	// Include these parameters if a specific implementation for the Start and Shutdown function are not needed
	component.StartFunc
	component.ShutdownFunc
}

// newConnector is a function to create a new connector
func newConnector(logger *zap.Logger, config component.Config) (*connectorImp, error) {
	logger.Info("Building spanhostmetrics connector")
	cfg := config.(*Config)

	return &connectorImp{
		config: *cfg,
		logger: logger,
	}, nil
}

// Capabilities implements the consumer interface.
func (c *connectorImp) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// ConsumeTraces method is called for each instance of a trace sent to the connector
func (c *connectorImp) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	// loop through the levels of spans of the one trace consumed
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		resourceSpan := td.ResourceSpans().At(i)

		for j := 0; j < resourceSpan.ScopeSpans().Len(); j++ {
			attrs := resourceSpan.Resource().Attributes()
			mapping := attrs.AsRaw()

			for key, v := range mapping {
				for _, attrName := range c.config.AttributeNames {
					if key == attrName {
						// create metric only if span of trace has one of the specific attributes
						metrics := pmetric.NewMetrics()
						ilm := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
						ilm.Scope().SetName("spanhostmetricsconnector")
						m := ilm.Metrics().AppendEmpty()
						m.SetName("span_host_info")
						m.SetEmptyGauge()
						dps := m.Gauge().DataPoints()
						dps.EnsureCapacity(1)
						timestamp := pcommon.NewTimestampFromTime(time.Now())
						dpCalls := dps.AppendEmpty()
						dpCalls.SetStartTimestamp(timestamp)
						dpCalls.SetTimestamp(timestamp)
						dpCalls.Attributes().PutStr("hostname", v.(string))
						dpCalls.SetIntValue(int64(1))

						// TODO: store the metric in a map and flush it after a certain time instead of on every span
						err := c.metricsConsumer.ConsumeMetrics(ctx, metrics)
						if err != nil {
							return err
						}
						break
					}
				}
			}
		}
	}
	return nil
}
