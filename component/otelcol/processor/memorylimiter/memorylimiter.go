// Package memorylimiter provides an otelcol.processor.memory_limiter component.
package memorylimiter

import (
	"fmt"
	"time"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.processor.memory_limiter",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := memorylimiterprocessor.NewFactory()
			return processor.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.processor.memory_limiter component.
type Arguments struct {
	CheckInterval         time.Duration    `river:"check_interval,attr"`
	MemoryLimit           units.Base2Bytes `river:"limit,attr,optional"`
	MemorySpikeLimit      units.Base2Bytes `river:"spike_limit,attr,optional"`
	MemoryLimitPercentage uint32           `river:"limit_percentage,attr,optional"`
	MemorySpikePercentage uint32           `river:"spike_limit_percentage,attr,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ processor.Arguments = Arguments{}
	_ river.Unmarshaler   = (*Arguments)(nil)
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	CheckInterval:         0,
	MemoryLimit:           0,
	MemorySpikeLimit:      0,
	MemoryLimitPercentage: 0,
	MemorySpikePercentage: 0,
}

// UnmarshalRiver implements river.Unmarshaler. It applies defaults to args and
// validates settings provided by the user.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}

	if args.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be greater than zero")
	}

	if args.MemoryLimit > 0 && args.MemoryLimitPercentage > 0 {
		return fmt.Errorf("either limit or limit_percentage must be set, but not both")
	}

	if args.MemoryLimit > 0 {
		if args.MemorySpikeLimit >= args.MemoryLimit {
			return fmt.Errorf("spike_limit must be less than limit")
		}
		if args.MemorySpikeLimit == 0 {
			args.MemorySpikeLimit = args.MemoryLimit / 5
		}
		return nil
	}
	if args.MemoryLimitPercentage > 0 {
		if args.MemoryLimitPercentage <= 0 ||
			args.MemoryLimitPercentage > 100 ||
			args.MemorySpikePercentage <= 0 ||
			args.MemorySpikePercentage > 100 {

			return fmt.Errorf("limit_percentage and spike_limit_percentage must be greater than 0 and and less or equal than 100")
		}
		return nil
	}

	return fmt.Errorf("either limit or limit_percentage must be set to greater than zero")
}

// Convert implements processor.Arguments.
func (args Arguments) Convert() (otelconfig.Processor, error) {
	return &memorylimiterprocessor.Config{
		ProcessorSettings: otelconfig.NewProcessorSettings(otelconfig.NewComponentID("memory_limiter")),

		CheckInterval:         args.CheckInterval,
		MemoryLimitMiB:        uint32(args.MemoryLimit / units.Mebibyte),
		MemorySpikeLimitMiB:   uint32(args.MemorySpikeLimit / units.Mebibyte),
		MemoryLimitPercentage: args.MemoryLimitPercentage,
		MemorySpikePercentage: args.MemorySpikePercentage,
	}, nil
}

// Extensions implements processor.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements processor.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

// NextConsumers implements processor.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}
