package integrations

import (
	"fmt"
	"reflect"

	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/util"
)

var (
	registeredIntegrations = []Config{}
	configFieldNames       = make(map[reflect.Type]string)

	emptyStructType = reflect.TypeOf(struct{}{})
	configsType     = reflect.TypeOf(Configs{})
)

// RegisterIntegration dynamically registers a new integration. The Config
// will represent the configuration that controls the specific integration.
// Registered Configs may be loaded using UnmarshalYAML or manually
// constructed.
//
// RegisterIntegration panics if cfg is not a pointer.
func RegisterIntegration(cfg Config) {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("RegisterIntegration must be given a pointer, got %T", cfg))
	}
	registeredIntegrations = append(registeredIntegrations, cfg)
	configFieldNames[reflect.TypeOf(cfg)] = cfg.Name()
}

// RegisteredIntegrations all Configs that were passed to RegisterIntegration.
// Each call will generate a new set of pointers.
func RegisteredIntegrations() []Config {
	res := make([]Config, 0, len(registeredIntegrations))
	for _, in := range registeredIntegrations {
		res = append(res, cloneIntegration(in))
	}
	return res
}

func cloneIntegration(c Config) Config {
	return reflect.New(reflect.TypeOf(c).Elem()).Interface().(Config)
}

// Configs is a list of UnmarshaledConfig. Configs for integrations which are
// unmarshaled from YAML are combined with common settings.
type Configs []UnmarshaledConfig

// MarshalYAML implements yaml.Marshaler for the configs in cc, marshaling to a
// YAML object.
func (cc Configs) MarshalYAML() (interface{}, error) {
	configTypes := make([]Config, 0, len(cc))
	for _, c := range cc {
		configTypes = append(configTypes, c.Config)
	}

	var (
		structType    = getConfigTypeForIntegrations(configTypes)
		structPointer = reflect.New(structType)
		structVal     = structPointer.Elem()
	)

	for _, c := range cc {
		field := structVal.FieldByName("Config_" + c.Name())

		// Marshal the common settings and config into raw YAML.
		bb, err := util.MarshalYAMLMerged(c.Common, c.Config)
		if err != nil {
			return nil, err
		}
		raw := util.RawYAML(bb)
		field.Set(reflect.ValueOf(&raw))
	}

	return structPointer.Interface(), nil
}

// UnmarshaledConfig combines an integration's config with common settings.
type UnmarshaledConfig struct {
	Config
	Common config.Common
}

// UnmarshalIntegrations unmarshals from the set of registered integrations,
// returning the set of loaded configs.
func UnmarshalIntegrations(unmarshal func(interface{}) error, registered []Config) (Configs, error) {
	var (
		structType    = getConfigTypeForIntegrations(registered)
		structPointer = reflect.New(structType)
		structVal     = structPointer.Elem()
	)

	// Unmarshal into our dynamic type.
	if err := unmarshal(structPointer.Interface()); err != nil {
		return nil, err
	}

	// Iterate through all of our fields and
	// structs where the first field is a config.Common and the second field is
	// a Config.
	var configs Configs
	for i, reg := range registered {
		// Our integrations are unmarshaled as *util.RawYAML. If it's nil, we treat
		// it as not present.
		raw, _ := structVal.Field(i).Interface().(*util.RawYAML)
		if raw == nil {
			continue
		}

		// Unmarshal into both our config instance and the common settings.
		var (
			cfg    = cloneIntegration(reg)
			common config.Common
		)
		err := util.UnmarshalYAMLMerged(*raw, &common, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal integration %q: %w", cfg.Name(), err)
		}

		configs = append(configs, UnmarshaledConfig{Config: cfg, Common: common})
	}

	return configs, nil
}

// getConfigTypeForIntegrations returns a dynamic struct type that has one
// field per provided integration.
//
// integrations are unmarshaled to *util.RawYAML for deferred unmarshaling.
func getConfigTypeForIntegrations(integrations []Config) reflect.Type {
	var fields []reflect.StructField
	for _, cfg := range integrations {
		fieldName := "Config_" + cfg.Name()
		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s,omitempty"`, cfg.Name())),
			Type: reflect.PtrTo(reflect.TypeOf(util.RawYAML{})),
		})
	}
	return reflect.StructOf(fields)
}
