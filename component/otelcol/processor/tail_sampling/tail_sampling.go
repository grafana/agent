// Package tail_sampling provides an otelcol.processor.tail_sampling component.
package tail_sampling

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	"github.com/grafana/agent/pkg/river"
	tsp "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.processor.tail_sampling",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := tsp.NewFactory()
			return processor.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.processor.tail_sampling component.
type Arguments struct {
	PolicyCfgs              []PolicyCfg   `river:"policy,block"`
	DecisionWait            time.Duration `river:"decision_wait,attr,optional"`
	NumTraces               uint64        `river:"num_traces,attr,optional"`
	ExpectedNewTracesPerSec uint64        `river:"expected_new_traces_per_sec,attr,optional"`
	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ processor.Arguments = Arguments{}
	_ river.Unmarshaler   = (*Arguments)(nil)
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	DecisionWait:            30 * time.Second,
	NumTraces:               50000,
	ExpectedNewTracesPerSec: 0,
}

// UnmarshalRiver implements river.Unmarshaler. It applies defaults to args and
// validates settings provided by the user.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}

	if args.DecisionWait.Milliseconds() <= 0 {
		return fmt.Errorf("decision_wait must be greater than zero")
	}

	if args.NumTraces <= 0 {
		return fmt.Errorf("num_traces must be greater than zero")
	}

	return nil
}

// Convert implements processor.Arguments.
func (args Arguments) Convert() otelcomponent.Config {
	var otelPolicyCfgs []tsp.PolicyCfg
	for _, policyCfg := range args.PolicyCfgs {
		otelPolicyCfgs = append(otelPolicyCfgs, policyCfg.Convert())
	}

	// TODO: is this still needed?
	/*
		mustDecodeMapStructure(map[string]interface{}{
			"decision_wait":               args.DecisionWait,
			"num_traces":                  args.NumTraces,
			"expected_new_traces_per_sec": args.ExpectedNewTracesPerSec,
			"policies":                    otelPolicyCfgs,
		}, &otelConfig)

		return &otelConfig
	*/

	return tsp.Config{
		DecisionWait:            args.DecisionWait,
		NumTraces:               args.NumTraces,
		ExpectedNewTracesPerSec: args.ExpectedNewTracesPerSec,
		PolicyCfgs:              otelPolicyCfgs,
	}
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
