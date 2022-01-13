package integrations

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	v1 "github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/util"
)

var (
	integrationNames       = make(map[string]interface{})  // Cache of field names for uniqueness checking.
	configFieldNames       = make(map[reflect.Type]string) // Map of registered type to field name
	integrationTypes       = make(map[reflect.Type]Type)   // Map of registered type to Type
	integrationTypesByName = make(map[string]Type)         // Map of name to Type

	// Registered integrations. Registered integrations may be either a Config or
	// a v1.Config. v1.Configs must have a corresponding upgrader for their type.
	registered = map[string]func() interface{}{}
	upgraders  = make(map[string]UpgradeFunc)

	emptyStructType = reflect.TypeOf(struct{}{})
	configsType     = reflect.TypeOf(Configs{})
)

// Register dynamically registers a new integration. The Config
// will represent the configuration that controls the specific integration.
// Registered Configs may be loaded using UnmarshalYAML or manually
// constructed.
//
// ty controls how the integration can be unmarshaled from YAML.
//
// Register panics if cfg is not a pointer.
func Register(f func() interface{}, ty Type) {
	cfg := f().(Config)
	registerIntegration(f, cfg.Name(), ty, nil)
}

func registerIntegration(f func() interface{}, name string, ty Type, upgrader UpgradeFunc) {
	cfg := f()
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Register must be given a pointer, got %T", cfg))
	}
	if _, exist := integrationNames[name]; exist {
		panic(fmt.Sprintf("Integration %q registered twice", name))
	}
	integrationNames[name] = name

	registered[name] = f

	configTy := reflect.TypeOf(cfg)
	integrationTypes[configTy] = ty
	configFieldNames[configTy] = name
	integrationTypesByName[name] = ty
	upgraders[name] = upgrader
}

// RegisterLegacy registers a v1.Config. upgrader will be used to upgrade it.
// upgrader will only be invoked after unmarshaling cfg from YAML, and the
// upgraded Config will be unwrapped again when marshaling back to YAML.
//
// RegisterLegacy only exists for the transition period where the v2
// integrations subsystem is an experiment. RegisterLegacy will be removed at a
// later date.
func RegisterLegacy(f func() interface{}, ty Type, upgrader UpgradeFunc) {
	cfg := f().(v1.Config)
	registerIntegration(f, cfg.Name(), ty, upgrader)
}

// UpgradeFunc upgrades cfg to a UpgradedConfig.
type UpgradeFunc func(cfg v1.Config, common common.MetricsConfig) UpgradedConfig

// UpgradedConfig is a v2 Config that was constructed through a legacy
// v1.Config. It allows unwrapping to retrieve the original config for
// the purposes of marshaling or unmarshaling.
type UpgradedConfig interface {
	Config

	// LegacyConfig returns the old v1.Config.
	LegacyConfig() (v1.Config, common.MetricsConfig)
}

// Type determines a specific type of integration.
type Type int

const (
	// TypeSingleton is an integration that can only be defined exactly once in
	// the config, unmarshaled through "<integration name>"
	TypeSingleton Type = iota

	// TypeMultiplex is an integration that can only be defined through an array,
	// unmarshaled through "<integration name>_configs"
	TypeMultiplex

	// TypeEither is an integration that can be unmarshaled either as a singleton
	// or as an array, but not both.
	//
	// Deprecated. Use either TypeSingleton or TypeMultiplex for new integrations.
	TypeEither
)

// setRegistered is used by tests to temporarily set integrations. Registered
// integrations will be unregistered after the test completes.
//
// setRegistered must not be used with parallelized tests.
func setRegistered(t *testing.T, cc map[Config]Type) {
	clear := func() {
		integrationNames = make(map[string]interface{})
		integrationTypes = make(map[reflect.Type]Type)
		integrationTypesByName = make(map[string]Type)
		configFieldNames = make(map[reflect.Type]string)
		registered = make(map[string]func() interface{})
		upgraders = make(map[string]UpgradeFunc)
	}

	t.Cleanup(clear)
	clear()

	for c, t := range cc {
		Register(func() interface{} { return c }, t)
	}
}

// Registered all Configs that were passed to Register or RegisterLegacy. Each
// call will generate a new set of configs.
func Registered() []Config {
	res := make([]Config, 0, len(registered))
	for _, r := range registered {
		res = append(res, cloneConfig(r))
	}
	return res
}

func cloneConfig(r func() interface{}) Config {
	c := r()
	switch v := c.(type) {
	case Config:
		return v
	case v1.Config:
		mut, ok := upgraders[v.Name()]
		if !ok || mut == nil {
			panic(fmt.Sprintf("Could not find transformer for legacy integration %T", r))
		}
		return mut(v, common.MetricsConfig{})
	default:
		panic(fmt.Sprintf("unexpected type %T", r))
	}
}

// Configs is a list of integrations. Note that Configs does not implement
// yaml.Unmarshaler or yaml.Marshaler. Use the UnmarshalYAML or MarshalYAML
// methods to deal with integrations defined from YAML.
type Configs []Config

// MarshalYAML helps implement yaml.Marshaler for structs that have a Configs
// field that should be inlined in the YAML string.
func MarshalYAML(v interface{}) (interface{}, error) {
	inVal := reflect.ValueOf(v)
	for inVal.Kind() == reflect.Ptr {
		inVal = inVal.Elem()
	}
	if inVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("integrations: can only marshal a struct, got %T", v)
	}
	inType := inVal.Type()

	var (
		outType    = getConfigTypeForIntegrations(inType)
		outPointer = reflect.New(outType)
		outVal     = outPointer.Elem()
	)

	// Copy over any existing value from inVal to cfgVal.
	//
	// The ordering of fields in inVal and cfgVal match identically up until the
	// extra fields appended to the end of cfgVal.
	var configs Configs
	for i, n := 0, inType.NumField(); i < n; i++ {
		if inType.Field(i).Type == configsType {
			configs = inVal.Field(i).Interface().(Configs)
			if configs == nil {
				configs = Configs{}
			}
		}
		if outType.Field(i).PkgPath != "" {
			continue // Field is unexported: ignore.
		}
		outVal.Field(i).Set(inVal.Field(i))
	}
	if configs == nil {
		return nil, fmt.Errorf("integrations: Configs field not found in type: %T", v)
	}

	for _, c := range configs {
		fieldName := c.Name()

		var data interface{} = c
		if wc, ok := c.(UpgradedConfig); ok {
			data, _ = wc.LegacyConfig()
		}

		integrationType, ok := integrationTypes[reflect.TypeOf(data)]
		if !ok {
			panic(fmt.Sprintf("config not registered: %T", data))
		}

		// Generate the *util.RawYAML to marshal out with.
		var (
			bb  []byte
			err error
		)
		switch v := c.(type) {
		case UpgradedConfig:
			inner, common := v.LegacyConfig()
			bb, err = util.MarshalYAMLMerged(common, inner)
		default:
			bb, err = yaml.Marshal(v)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to marshal integration %q: %w", fieldName, err)
		}
		raw := util.RawYAML(bb)

		switch integrationType {
		case TypeSingleton:
			field := outVal.FieldByName("XXX_Config_" + fieldName)
			field.Set(reflect.ValueOf(&raw))
		case TypeMultiplex, TypeEither:
			field := outVal.FieldByName("XXX_Configs_" + fieldName)
			field.Set(reflect.Append(field, reflect.ValueOf(&raw)))
		}
	}

	return outPointer.Interface(), nil
}

// UnmarshalYAML helps implement yaml.Unmarshaller for structs that have a
// Configs field that should be inlined in the YAML string. Code adapted from
// Prometheus:
//
//   https://github.com/prometheus/prometheus/blob/511511324adfc4f4178f064cc104c2deac3335de/discovery/registry.go#L111
func UnmarshalYAML(out interface{}, unmarshal func(interface{}) error) error {
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
		cfgType    = getConfigTypeForIntegrations(outType)
		cfgPointer = reflect.New(cfgType)
		cfgVal     = cfgPointer.Elem()
	)

	// Copy over any existing value from outVal to cfgVal.
	//
	// The ordering of fields in outVal and cfgVal match identically up until the
	// extra fields appended to the end of cfgVal.
	var configs *Configs
	for i := 0; i < outVal.NumField(); i++ {
		if outType.Field(i).Type == configsType {
			if configs != nil {
				return fmt.Errorf("integrations: Multiple Configs fields found in %T", out)
			}
			configs = outVal.Field(i).Addr().Interface().(*Configs)
			continue
		}
		if cfgType.Field(i).PkgPath != "" {
			// Ignore unexported fields
			continue
		}
		cfgVal.Field(i).Set(outVal.Field(i))
	}
	if configs == nil {
		return fmt.Errorf("integrations: No Configs field found in %T", out)
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

	// Iterate through the remainder of our fields, which should all be
	// either a Config or a slice of types that implement Config.
	for i := outVal.NumField(); i < cfgVal.NumField(); i++ {
		// Our integrations are unmarshaled as *util.RawYAML or []*util.RawYAML. If
		// it's nil, we treat it as not defined.
		fieldType := cfgVal.Type().Field(i)
		field := cfgVal.Field(i)
		if field.IsNil() {
			continue
		}

		switch field.Kind() {
		case reflect.Slice:
			configName := strings.TrimPrefix(fieldType.Name, "XXX_Configs_")
			configFunc, ok := registered[configName]
			if !ok {
				return fmt.Errorf("integration %q not registered", configName)
			}
			configReference := configFunc()
			for i := 0; i < field.Len(); i++ {
				if field.Index(i).IsNil() {
					continue
				}
				raw := field.Index(i).Interface().(*util.RawYAML)
				c, err := deferredConfigUnmarshal(*raw, configReference)
				if err != nil {
					return err
				}
				*configs = append(*configs, c)
			}
		default:
			configName := strings.TrimPrefix(fieldType.Name, "XXX_Config_")
			_, ok := integrationNames[configName]
			if !ok {
				return fmt.Errorf("integration %q not registered", configName)
			}
			configFunc, ok := registered[configName]
			if !ok {
				return fmt.Errorf("integration %q not registered", configName)
			}
			raw := field.Interface().(*util.RawYAML)
			configReference := configFunc()
			c, err := deferredConfigUnmarshal(*raw, configReference)
			if err != nil {
				return err
			}
			*configs = append(*configs, c)
		}
	}

	return nil
}

// deferredConfigUnmarshal performs a deferred unmarshal of raw into a Config.
// ref must be either Config or v1.Config.
func deferredConfigUnmarshal(raw util.RawYAML, ref interface{}) (Config, error) {
	switch ref := ref.(type) {
	case Config:
		out := ref
		err := yaml.Unmarshal(raw, out)
		return out, err
	case v1.Config:
		mut, ok := upgraders[ref.Name()]
		if !ok {
			panic(fmt.Sprintf("unexpected type %T", ref))
		}
		var (
			common common.MetricsConfig
			out    = ref
		)
		err := util.UnmarshalYAMLMerged(raw, &common, out)
		return mut(out, common), err
	default:
		panic(fmt.Sprintf("unexpected type %T", ref))
	}
}

// getConfigTypeForIntegrations returns a dynamic struct type that has all of
// the same fields as out including the fields for the provided integrations.
//
// integrations are unmarshaled to *util.RawYAML for deferred unmarshaling.
func getConfigTypeForIntegrations(out reflect.Type) reflect.Type {
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

	for _, reg := range registered {
		// Fields use a prefix that's unlikely to collide with anything else.
		c := reg()
		configTy := reflect.TypeOf(c)
		integrationType := integrationTypes[configTy]
		fieldName := configFieldNames[configTy]

		singletonType := reflect.PtrTo(reflect.TypeOf(util.RawYAML{}))

		if integrationType == TypeSingleton || integrationType == TypeEither {
			fields = append(fields, reflect.StructField{
				Name: "XXX_Config_" + fieldName,
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s,omitempty"`, fieldName)),
				Type: singletonType,
			})
		}
		if integrationType == TypeMultiplex || integrationType == TypeEither {
			fields = append(fields, reflect.StructField{
				Name: "XXX_Configs_" + fieldName,
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s_configs,omitempty"`, fieldName)),
				Type: reflect.SliceOf(singletonType),
			})
		}
	}
	return reflect.StructOf(fields)
}

func replaceYAMLTypeError(err error, oldTyp, newTyp reflect.Type) error {
	if e, ok := err.(*yaml.TypeError); ok {
		oldStr := oldTyp.String()
		newStr := newTyp.String()
		for i, s := range e.Errors {
			e.Errors[i] = strings.ReplaceAll(s, oldStr, newStr)
		}
	}
	return err
}

// UnmarshalYamlToExporters attempts to convert the contents of yaml string into a set of exporters and then return
// those configurations.
func UnmarshalYamlToExporters(contents string) ([]Config, error) {
	unmarshalledConfigs := make([]Config, 0)
	// Registered returns a cloned version of the config, so we are fine in using it
	for _, cfg := range Registered() {
		integrationType := integrationTypesByName[cfg.Name()]
		var fields []reflect.StructField
		// Integrations can either have multiple copies (multiplex ie redis) or singleton (windows)
		if integrationType == TypeMultiplex {
			fields = append(fields, reflect.StructField{
				Name: strings.ToUpper(cfg.Name()),
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s_configs,omitempty"`, cfg.Name())),
				Type: reflect.SliceOf(reflect.TypeOf(util.RawYAML{})),
			})
		} else {
			fields = append(fields, reflect.StructField{
				Name: strings.ToUpper(cfg.Name()),
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s,omitempty"`, cfg.Name())),
				Type: reflect.TypeOf(util.RawYAML{}),
			})
		}
		createdType := reflect.StructOf(fields)
		instanceElement := reflect.New(createdType).Elem()
		instance := instanceElement.Addr().Interface()
		if instance == nil {
			return nil, errors.New(fmt.Sprintf("unable to create config of name %s", cfg.Name()))
		}
		err := yaml.Unmarshal([]byte(contents), instance)
		if err != nil {
			return nil, err
		}
		exportersYaml := make([]util.RawYAML, 0)
		if integrationType == TypeMultiplex {
			exportersYaml = append(exportersYaml, instanceElement.Field(0).Interface().([]util.RawYAML)...)
		} else {
			exportersYaml = append(exportersYaml, instanceElement.Field(0).Interface().(util.RawYAML))
		}

		for _, exporterYaml := range exportersYaml {
			if exporterYaml != nil {
				f := registered[cfg.Name()]
				newCfg := f()
				shim := &ConfigShim{}
				shim.Orig = newCfg.(v1.Config)
				// The unmarshal needs to happen twice since passing in just the shim doesn't work due to the yaml tags.
				// We could probably simplify this by adding tags dynamically but since this is all in memory the speed
				// is likely acceptable for simplicity
				err = yaml.Unmarshal(exporterYaml, newCfg)
				if err != nil {
					return nil, err
				}
				err = yaml.Unmarshal(exporterYaml, &shim.Common)
				if err != nil {
					return nil, err
				}
				unmarshalledConfigs = append(unmarshalledConfigs, shim)
			}

		}
	}
	return unmarshalledConfigs, nil
}

type MarshallingConfig struct {
	Config
}
