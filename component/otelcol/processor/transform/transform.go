// Package transform provides an otelcol.processor.transform component.
package transform

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.processor.transform",
		Args:          Arguments{},
		Exports:       otelcol.ConsumerExports{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := transformprocessor.NewFactory()
			return processor.New(opts, fact, args.(Arguments))
		},
	})
}

type ContextID string

const (
	Resource  ContextID = "resource"
	Scope     ContextID = "scope"
	Span      ContextID = "span"
	SpanEvent ContextID = "spanevent"
	Metric    ContextID = "metric"
	DataPoint ContextID = "datapoint"
	Log       ContextID = "log"
)

func (c *ContextID) UnmarshalText(text []byte) error {
	str := ContextID(strings.ToLower(string(text)))
	switch str {
	case Resource, Scope, Span, SpanEvent, Metric, DataPoint, Log:
		*c = str
		return nil
	default:
		return fmt.Errorf("unknown context %v", str)
	}
}

type contextStatementsSlice []contextStatements

type contextStatements struct {
	Context    ContextID `river:"context,attr"`
	Statements []string  `river:"statements,attr"`
}

// Arguments configures the otelcol.processor.transform component.
type Arguments struct {
	// ErrorMode determines how the processor reacts to errors that occur while processing a statement.
	ErrorMode        ottl.ErrorMode         `river:"error_mode,attr,optional"`
	TraceStatements  contextStatementsSlice `river:"trace_statements,block,optional"`
	MetricStatements contextStatementsSlice `river:"metric_statements,block,optional"`
	LogStatements    contextStatementsSlice `river:"log_statements,block,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ processor.Arguments = Arguments{}
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	ErrorMode: ottl.PropagateError,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	otelArgs, err := args.convertImpl()
	if err != nil {
		return err
	}
	return otelArgs.Validate()
}

func (stmts *contextStatementsSlice) convert() []interface{} {
	if stmts == nil {
		return nil
	}

	res := make([]interface{}, 0, len(*stmts))

	if len(*stmts) == 0 {
		return res
	}

	for _, stmt := range *stmts {
		res = append(res, stmt.convert())
	}
	return res
}

func (args *contextStatements) convert() map[string]interface{} {
	if args == nil {
		return nil
	}

	return map[string]interface{}{
		"context":    args.Context,
		"statements": args.Statements,
	}
}

// Convert implements processor.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return args.convertImpl()
}

// convertImpl is a helper function which returns the real type of the config,
// instead of the otelcomponent.Config interface.
func (args Arguments) convertImpl() (*transformprocessor.Config, error) {
	input := make(map[string]interface{})

	input["error_mode"] = args.ErrorMode

	if len(args.TraceStatements) > 0 {
		input["trace_statements"] = args.TraceStatements.convert()
	}

	if len(args.MetricStatements) > 0 {
		input["metric_statements"] = args.MetricStatements.convert()
	}

	if len(args.LogStatements) > 0 {
		input["log_statements"] = args.LogStatements.convert()
	}

	var result transformprocessor.Config
	err := mapstructure.Decode(input, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Extensions implements processor.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements processor.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements processor.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}
