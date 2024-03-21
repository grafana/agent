package count

import (
	"fmt"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/connector"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.uber.org/multierr"
)

func init() {
	component.Register(component.Registration{
		Name:      "otelcol.connector.count",
		Stability: featuregate.StabilityExperimental,
		Args:      Arguments{},
		Exports:   otelcol.ConsumerExports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := countconnector.NewFactory()
			return connector.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.connector.count component.
type Arguments struct {
	Spans      []MetricInfo `river:"span,block,optional"`
	SpanEvents []MetricInfo `river:"spanevent,block,optional"`
	Metrics    []MetricInfo `river:"metric,block,optional"`
	DataPoints []MetricInfo `river:"datapoint,block,optional"`
	Logs       []MetricInfo `river:"log,block,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ river.Validator     = (*Arguments)(nil)
	_ river.Defaulter     = (*Arguments)(nil)
	_ connector.Arguments = (*Arguments)(nil)
)

// Default metrics are emitted if no conditions are specified.
const (
	defaultMetricNameSpans      = "trace.span.count"
	defaultMetricDescSpans      = "The number of spans observed."
	defaultMetricNameSpanEvents = "trace.span.event.count"
	defaultMetricDescSpanEvents = "The number of span events observed."

	defaultMetricNameMetrics    = "metric.count"
	defaultMetricDescMetrics    = "The number of metrics observed."
	defaultMetricNameDataPoints = "metric.datapoint.count"
	defaultMetricDescDataPoints = "The number of data points observed."

	defaultMetricNameLogs = "log.record.count"
	defaultMetricDescLogs = "The number of log records observed."
)

var DefaultArguments = Arguments{
	Spans: []MetricInfo{
		{
			Name:        defaultMetricNameSpans,
			Description: defaultMetricDescSpans,
		},
	},
	SpanEvents: []MetricInfo{
		{
			Name:        defaultMetricNameSpanEvents,
			Description: defaultMetricDescSpanEvents,
		},
	},
	Metrics: []MetricInfo{
		{
			Name:        defaultMetricNameMetrics,
			Description: defaultMetricDescMetrics,
		},
	},
	DataPoints: []MetricInfo{
		{
			Name:        defaultMetricNameDataPoints,
			Description: defaultMetricDescDataPoints,
		},
	},
	Logs: []MetricInfo{
		{
			Name:        defaultMetricNameLogs,
			Description: defaultMetricDescLogs,
		},
	},
}

// ConnectorType implements connector.Arguments.
func (Arguments) ConnectorType() connector.Type {
	return connector.ConnectorLogsToMetrics | connector.ConnectorTracesToMetrics | connector.ConnectorMetricsToMetrics
}

// Convert implements connector.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	var (
		spans      = convertMetricInfo(args.Spans)
		spanEvents = convertMetricInfo(args.SpanEvents)
		metrics    = convertMetricInfo(args.Metrics)
		dataPoints = convertMetricInfo(args.DataPoints)
		logs       = convertMetricInfo(args.Logs)
	)
	cfg := &countconnector.Config{
		Spans:      spans,
		SpanEvents: spanEvents,
		Metrics:    metrics,
		DataPoints: dataPoints,
		Logs:       logs,
	}
	if err := checkConverted(args, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func checkConverted(args Arguments, cfg *countconnector.Config) error {
	var err error
	err = multierr.Append(err, dupeCheck("span", args.Spans, cfg.Spans))
	err = multierr.Append(err, dupeCheck("spanevent", args.SpanEvents, cfg.SpanEvents))
	err = multierr.Append(err, dupeCheck("metric", args.Metrics, cfg.Metrics))
	err = multierr.Append(err, dupeCheck("datapoint", args.DataPoints, cfg.DataPoints))
	err = multierr.Append(err, dupeCheck("log", args.Logs, cfg.Logs))
	return err
}

func dupeCheck(signalType string, mi []MetricInfo, ccmi map[string]countconnector.MetricInfo) error {
	var err error
	if len(mi) != len(ccmi) {
		nameMap := make(map[string]struct{})
		for _, entry := range mi {
			if _, ok := nameMap[entry.Name]; ok {
				err = multierr.Append(err, fmt.Errorf("duplicate %s name: %s", signalType, entry.Name))
			} else {
				nameMap[entry.Name] = struct{}{}
			}
		}
	}
	return err
}

func convertMetricInfo(mi []MetricInfo) map[string]countconnector.MetricInfo {
	ret := make(map[string]countconnector.MetricInfo)
	for _, metricInfo := range mi {
		var attrConfigs []countconnector.AttributeConfig
		for _, ac := range metricInfo.Attributes {
			a := countconnector.AttributeConfig{
				Key:          ac.Key,
				DefaultValue: ac.DefaultValue,
			}
			attrConfigs = append(attrConfigs, a)
		}
		ret[metricInfo.Name] = countconnector.MetricInfo{
			Description: metricInfo.Description,
			Conditions:  metricInfo.Conditions,
			Attributes:  attrConfigs,
		}
	}
	return ret
}

// Exporters implements connector.Arguments.
func (Arguments) Exporters() map[otelcomponent.Type]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// Extensions implements connector.Arguments.
func (Arguments) Extensions() map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements connector.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	cfg, err := args.Convert()
	if err != nil {
		return err
	}
	return cfg.(*countconnector.Config).Validate()
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}
