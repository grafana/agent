package integrations

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

var (
	registeredIntegrations = []Config{}

	configFieldNames = make(map[reflect.Type]string)
	integrationTypes = make(map[reflect.Type]Type)

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
func Register(cfg Config, ty Type) {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Register must be given a pointer, got %T", cfg))
	}
	if _, exist := configFieldNames[reflect.TypeOf(cfg)]; exist {
		panic(fmt.Sprintf("Integration %T registered twice", cfg))
	}

	registeredIntegrations = append(registeredIntegrations, cfg)

	configTy := reflect.TypeOf(cfg)
	integrationTypes[configTy] = ty
	configFieldNames[configTy] = cfg.Name()
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
		registeredIntegrations = registeredIntegrations[:0]
		integrationTypes = make(map[reflect.Type]Type)
		configFieldNames = make(map[reflect.Type]string)
	}

	t.Cleanup(clear)
	clear()

	for c, t := range cc {
		Register(c, t)
	}
}

// Registered all Configs that were passed to Register.
// Each call will generate a new set of configs.
func Registered() []Config {
	res := make([]Config, 0, len(registeredIntegrations))
	for _, in := range registeredIntegrations {
		res = append(res, cloneIntegration(in))
	}
	return res
}

func cloneIntegration(c Config) Config {
	return reflect.New(reflect.TypeOf(c).Elem()).Interface().(Config)
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
		fieldName, ok := configFieldNames[reflect.TypeOf(c)]
		if !ok {
			return nil, fmt.Errorf("integrations: cannot marshal unregistered Config type: %T", c)
		}

		integrationType := integrationTypes[reflect.TypeOf(c)]
		switch integrationType {
		case TypeSingleton:
			field := outVal.FieldByName("XXX_Config_" + fieldName)
			field.Set(reflect.ValueOf(c))
		case TypeMultiplex, TypeEither:
			field := outVal.FieldByName("XXX_Configs_" + fieldName)
			field.Set(reflect.Append(field, reflect.ValueOf(c)))
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
		field := cfgVal.Field(i)

		if field.IsNil() {
			continue
		}

		switch field.Kind() {
		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				val := field.Index(i).Interface().(Config)
				*configs = append(*configs, val)
			}
		default:
			val := field.Interface().(Config)
			*configs = append(*configs, val)
		}
	}

	return nil
}

// getConfigTypeForIntegrations returns a dynamic struct type that has all of
// the same fields as out including the fields for the provided integrations.
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

	for _, cfg := range registeredIntegrations {
		// Fields use a prefix that's unlikely to collide with anything else.
		configTy := reflect.TypeOf(cfg)
		integrationType := integrationTypes[configTy]

		if integrationType == TypeSingleton || integrationType == TypeEither {
			fields = append(fields, reflect.StructField{
				Name: "XXX_Config_" + cfg.Name(),
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s,omitempty"`, cfg.Name())),
				Type: configTy,
			})
		}
		if integrationType == TypeMultiplex || integrationType == TypeEither {
			fields = append(fields, reflect.StructField{
				Name: "XXX_Configs_" + cfg.Name(),
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s_configs,omitempty"`, cfg.Name())),
				Type: reflect.SliceOf(configTy),
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
