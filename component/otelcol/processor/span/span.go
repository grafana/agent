// Package span provides an otelcol.processor.span component.
package span

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.processor.span",
		Args:          Arguments{},
		Exports:       otelcol.ConsumerExports{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := spanprocessor.NewFactory()
			return processor.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.processor.span component.
type Arguments struct {
	Match otelcol.MatchConfig `river:",squash"`

	// Name specifies the components required to re-name a span.
	Name Name `river:"name,block,optional"`

	// SetStatus specifies status which should be set for this span.
	SetStatus *Status `river:"status,block,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ processor.Arguments = Arguments{}
)

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args == nil {
		return nil
	}

	if args.SetStatus != nil {
		switch args.SetStatus.Code {
		case StatusCodeOk, StatusCodeError, StatusCodeUnset:
			// No error
		default:
			return fmt.Errorf("status code is set to an invalid value of %q", args.SetStatus.Code)
		}

		if args.SetStatus.Code != StatusCodeError && args.SetStatus.Description != "" {
			return fmt.Errorf("status description should be empty for non-error status codes")
		}
	}
	return nil
}

// Convert implements processor.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	input := make(map[string]interface{})

	if args.Match.Include != nil {
		matchConfig, err := args.Match.Include.Convert()
		if err != nil {
			return nil, fmt.Errorf("error getting 'include' match properties: %w", err)
		}
		if len(matchConfig) > 0 {
			input["include"] = matchConfig
		}
	}

	if args.Match.Exclude != nil {
		matchConfig, err := args.Match.Exclude.Convert()
		if err != nil {
			return nil, fmt.Errorf("error getting 'exclude' match properties: %w", err)
		}
		if len(matchConfig) > 0 {
			input["exclude"] = matchConfig
		}
	}

	var result spanprocessor.Config
	err := mapstructure.Decode(input, &result)

	if err != nil {
		return nil, err
	}

	result.Rename = *args.Name.Convert()
	result.SetStatus = args.SetStatus.Convert()

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

// Name specifies the attributes to use to re-name a span.
type Name struct {
	// Specifies transformations of span name to and from attributes.
	// First FromAttributes rules are applied, then ToAttributes are applied.
	// At least one of these 2 fields must be set.

	// FromAttributes represents the attribute keys to pull the values from to
	// generate the new span name. All attribute keys are required in the span
	// to re-name a span. If any attribute is missing from the span, no re-name
	// will occur.
	// Note: The new span name is constructed in order of the `from_attributes`
	// specified in the configuration. This field is required and cannot be empty.
	FromAttributes []string `river:"from_attributes,attr,optional"`

	// Separator is the string used to separate attributes values in the new
	// span name. If no value is set, no separator is used between attribute
	// values. Used with FromAttributes only.
	Separator string `river:"separator,attr,optional"`

	// ToAttributes specifies a configuration to extract attributes from span name.
	ToAttributes *ToAttributes `river:"to_attributes,block,optional"`
}

func (n *Name) Convert() *spanprocessor.Name {
	if n == nil {
		return nil
	}

	return &spanprocessor.Name{
		FromAttributes: n.FromAttributes,
		Separator:      n.Separator,
		ToAttributes:   n.ToAttributes.Convert(),
	}
}

// ToAttributes specifies a configuration to extract attributes from span name.
type ToAttributes struct {
	// Rules is a list of rules to extract attribute values from span name. The values
	// in the span name are replaced by extracted attribute names. Each rule in the list
	// is a regex pattern string. Span name is checked against the regex. If it matches
	// then all named subexpressions of the regex are extracted as attributes
	// and are added to the span. Each subexpression name becomes an attribute name and
	// subexpression matched portion becomes the attribute value. The matched portion
	// in the span name is replaced by extracted attribute name. If the attributes
	// already exist in the span then they will be overwritten. The process is repeated
	// for all rules in the order they are specified. Each subsequent rule works on the
	// span name that is the output after processing the previous rule.
	Rules []string `river:"rules,attr"`

	// BreakAfterMatch specifies if processing of rules should stop after the first
	// match. If it is false rule processing will continue to be performed over the
	// modified span name.
	BreakAfterMatch bool `river:"break_after_match,attr,optional"`
}

// DefaultArguments holds default settings for Arguments.
var DefaultToAttributes = ToAttributes{
	BreakAfterMatch: false,
}

// SetToDefault implements river.Defaulter.
func (args *ToAttributes) SetToDefault() {
	if args == nil {
		return
	}

	*args = DefaultToAttributes
}

func (ta *ToAttributes) Convert() *spanprocessor.ToAttributes {
	if ta == nil {
		return nil
	}

	return &spanprocessor.ToAttributes{
		Rules:           ta.Rules,
		BreakAfterMatch: ta.BreakAfterMatch,
	}
}

type Status struct {
	// Code is one of three values "Ok" or "Error" or "Unset". Please check:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#set-status
	Code string `river:"code,attr"`

	// Description is an optional field documenting Error statuses.
	Description string `river:"description,attr,optional"`
}

var (
	StatusCodeOk    = ptrace.StatusCodeOk.String()
	StatusCodeError = ptrace.StatusCodeError.String()
	StatusCodeUnset = ptrace.StatusCodeUnset.String()
)

func (s *Status) Convert() *spanprocessor.Status {
	if s == nil {
		return nil
	}

	return &spanprocessor.Status{
		Code:        s.Code,
		Description: s.Description,
	}
}
