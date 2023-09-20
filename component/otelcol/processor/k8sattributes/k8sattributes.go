// Package attributes provides an otelcol.processor.k8sattributes component.
package k8sattributes

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.processor.k8sattributes",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := k8sattributesprocessor.NewFactory()
			return processor.New(opts, fact, args.(Arguments))
		},
	})
}

var (
	_ processor.Arguments = Arguments{}
)

// Arguments configures the otelcol.processor.k8sattributes component.
type Arguments struct {
	AuthType        string              `river:"auth_type,attr,optional"`
	Passthrough     bool                `river:"passthrough,attr,optional"`
	ExtractConfig   ExtractConfig       `river:"extract,block,optional"`
	Filter          FilterConfig        `river:"filter,block,optional"`
	PodAssociations PodAssociationSlice `river:"pod_association,block,optional"`
	Exclude         ExcludeConfig       `river:"exclude,block,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	cfg, err := args.Convert()
	if err != nil {
		return err
	}

	return cfg.(*k8sattributesprocessor.Config).Validate()
}

// Convert implements processor.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	input := make(map[string]interface{})

	if args.AuthType == "" {
		input["auth_type"] = "serviceAccount"
	} else {
		input["auth_type"] = args.AuthType
	}

	input["passthrough"] = args.Passthrough

	if extract := args.ExtractConfig.convert(); len(extract) > 0 {
		input["extract"] = extract
	}

	if filter := args.Filter.convert(); len(filter) > 0 {
		input["filter"] = filter
	}

	if podAssociations := args.PodAssociations.convert(); len(podAssociations) > 0 {
		input["pod_association"] = podAssociations
	}

	if exclude := args.Exclude.convert(); len(exclude) > 0 {
		input["exclude"] = exclude
	}

	var result k8sattributesprocessor.Config
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
