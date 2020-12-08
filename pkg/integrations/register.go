package integrations

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations/common"
	"gopkg.in/yaml.v3"
)

var (
	registeredIntegrations = []IntegrationConfig{}

	emptyStructType = reflect.TypeOf(struct{}{})
	configsType     = reflect.TypeOf(IntegrationConfigs{})
)

// RegisterIntegration dynamically registers a new integration. The IntegrationConfig
// will represent the configuration that controls the specific integration.
// Registered IntegrationConfigs may be loaded using UnmarshalYAML or manually
// constructed.
//
// RegisterIntegration panics if cfg is not a pointer.
func RegisterIntegration(cfg IntegrationConfig) {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("RegisterIntegration must be given a pointer, got %T", cfg))
	}
	registeredIntegrations = append(registeredIntegrations, cfg)
}

// IntegrationConfig provides the configuration and constructor for an
// integration.
type IntegrationConfig interface {
	// Name returns the name of the integration and the key that will be used to
	// pull the configuration from the Agent config YAML.
	Name() string

	// IsEnabled returns whether this integration should run.
	IsEnabled() bool

	// NewIntegration returns an integration for the given with the given logger.
	NewIntegration(l log.Logger) (common.Integration, error)
}

// IntegrationConfigs is a list of integrations.
type IntegrationConfigs []IntegrationConfig

func (c *IntegrationConfigs) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return c.unmarshalWithIntegrations(registeredIntegrations, unmarshal)
}

func (c *IntegrationConfigs) unmarshalWithIntegrations(integrations []IntegrationConfig, unmarshal func(interface{}) error) error {
	// Create a dynamic struct type full of our registered integrations and
	// unmarshal to it.
	var fields []reflect.StructField
	for _, cfg := range integrations {
		fields = append(fields, reflect.StructField{
			Name: "Config_" + cfg.Name(),
			Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s"`, cfg.Name())),
			Type: reflect.TypeOf(cfg),
		})
	}

	var (
		structType = reflect.StructOf(fields)
		structVal  = reflect.New(structType)
	)
	if err := unmarshal(structVal.Interface()); err != nil {
		return err
	}

	// Go over all non-nil fields in structVal and append them to c.
	structVal = structVal.Elem()
	for i := 0; i < structVal.NumField(); i++ {
		if structVal.Field(i).IsNil() {
			continue
		}

		val := structVal.Field(i).Interface().(IntegrationConfig)
		*c = append(*c, val)
	}

	return nil
}

// UnmarshalYAML helps implement yaml.Unmarshaller for structs that have an
// IntegrationConfigs field that should be inlined in the YAML string.
func UnmarshalYAML(out interface{}, unmarshal func(interface{}) error) error {
	return unmarshalIntegrationsWithList(registeredIntegrations, out, unmarshal)
}

// unmarshalIntegrationsWithList unmarshals to a subtype of out that has a
// field added for every integration in integrations. Code adapted from
// Prometheus:
//
//   https://github.com/prometheus/prometheus/blob/511511324adfc4f4178f064cc104c2deac3335de/discovery/registry.go#L111
func unmarshalIntegrationsWithList(integrations []IntegrationConfig, out interface{}, unmarshal func(interface{}) error) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("integrations: can only unmarshal into a struct pointer, got %T", out)
	}
	outVal = outVal.Elem()
	if outVal.Kind() != reflect.Struct {
		return fmt.Errorf("integrations: can only unmarshal into a struct pointer, got %T", out)
	}
	outType := outVal.Type()

	var (
		cfgType    = getConfigTypeForIntegrations(integrations, outType)
		cfgPointer = reflect.New(cfgType)
		cfgVal     = cfgPointer.Elem()
	)

	// Copy over any existing value from outVal to cfgVal.
	//
	// The ordering of fields in outVal and cfgVal match identically up until the
	// extra fields appended to the end of cfgVal.
	var configs *IntegrationConfigs
	for i := 0; i < outVal.NumField(); i++ {
		if outType.Field(i).Type == configsType {
			if configs != nil {
				return fmt.Errorf("integrations: Multiple IntegrationConfigs fields found in %T", out)
			}
			configs = outVal.Field(i).Addr().Interface().(*IntegrationConfigs)
			continue
		}
		if cfgType.Field(i).PkgPath != "" {
			// Ignore unexported fields
			continue
		}
		cfgVal.Field(i).Set(outVal.Field(i))
	}
	if configs == nil {
		return fmt.Errorf("integrations: No IntegrationConfigs field found in %T", out)
	}

	// Unmarshal into our dynamic type.
	if err := unmarshal(cfgPointer.Interface()); err != nil {
		return replaceYAMLTypeError(err, cfgType, outType)
	}

	// Copy back unmarshaled fields that were originally in outVal.
	for i := 0; i < outVal.NumField(); i++ {
		if cfgType.Field(i).PkgPath != "" {
			// Ignore unexported fields
			continue
		}
		outVal.Field(i).Set(cfgVal.Field(i))
	}

	// Iterate through the remainder of our fields, which should all be of
	// type IntegrationConfig.
	for i := outVal.NumField(); i < cfgVal.NumField(); i++ {
		field := cfgVal.Field(i)

		if field.IsNil() {
			continue
		}
		val := cfgVal.Field(i).Interface().(IntegrationConfig)
		*configs = append(*configs, val)
	}

	return nil
}

// getConfigTypeForIntegrations returns a dynamic struct type that has all of
// the same fields as out including the fields for the provided integrations.
func getConfigTypeForIntegrations(integrations []IntegrationConfig, out reflect.Type) reflect.Type {
	// Initial exported fields map one-to-one.
	var fields []reflect.StructField
	for i, n := 0, out.NumField(); i < n; i++ {
		switch field := out.Field(i); {
		case field.PkgPath == "" && field.Type != configsType:
			fields = append(fields, field)
		default:
			fields = append(fields, reflect.StructField{
				Name:    "_" + field.Name, // Field must be unexported.
				PkgPath: out.PkgPath(),
				Type:    emptyStructType,
			})
		}
	}
	for _, cfg := range integrations {
		// Use a prefix that's unlikely to collide with anything else.
		fieldName := "XXX_Config_" + cfg.Name()
		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s"`, cfg.Name())),
			Type: reflect.TypeOf(cfg),
		})
	}
	return reflect.StructOf(fields)
}

func replaceYAMLTypeError(err error, oldTyp, newTyp reflect.Type) error {
	if e, ok := err.(*yaml.TypeError); ok {
		oldStr := oldTyp.String()
		newStr := newTyp.String()
		for i, s := range e.Errors {
			e.Errors[i] = strings.Replace(s, oldStr, newStr, -1)
		}
	}
	return err
}
