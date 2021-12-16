package integrations

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

var (
	integrationNames = make(map[string]struct{})     // Cache of field names for uniqueness checking.
	configFieldNames = make(map[reflect.Type]string) // Map of registered type to field name
	integrationTypes = make(map[reflect.Type]Type)   // Map of registered type to Type

	// Registered integrations. Registered integrations may be any type. If they
	// do not implement Config, then they must have a mutator to wrap them into
	// a Config.
	registered = []interface{}{}
	mutators   = make(map[reflect.Type]WrapperFunc)

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
	// Special case: we don't need a transformer because cfg is already a Config.
	RegisterDynamic(cfg, cfg.Name(), ty, nil)
}

// RegisterDynamic registers a dynamic integration v which does not implement
// Config directly. The transform function will be invoked to create a wrapper
// implementation of Config. transform will be invoked after unmarshaling v from
// YAML, and the result will be unwrapped back to the original form at marshal
// time.
//
// RegisterDynamic is only useful in niche instances. When possible, use
// Register instead.
//
// name is the name of the integration used for unmarshaling from YAML. Must be
// unique across all integrations.
func RegisterDynamic(v interface{}, name string, ty Type, transform WrapperFunc) {
	if _, isConfig := v.(Config); !isConfig && transform == nil {
		panic(fmt.Sprintf("transform must not be nil; %T does not implement Config", v))
	}
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("RegisterDynamic must be given a pointer, got %T", v))
	}
	if _, exist := integrationNames[name]; exist {
		panic(fmt.Sprintf("Integration %q registered twice", name))
	}
	integrationNames[name] = struct{}{}

	registered = append(registered, v)

	configTy := reflect.TypeOf(v)
	integrationTypes[configTy] = ty
	configFieldNames[configTy] = name
	mutators[configTy] = transform
}

// WrapperFunc wraps in in a WrappedConfig container.
type WrapperFunc func(in interface{}) WrappedConfig

// WrappedConfig represents a Config which is a container for some data. This
// is used when registering integrations through RegisterDynamic.
type WrappedConfig interface {
	Config

	// UnwrapConfig returns the inner data. The inner data is used for YAML
	// marshaling and unmarshaling.
	UnwrapConfig() interface{}
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
		integrationNames = make(map[string]struct{})
		integrationTypes = make(map[reflect.Type]Type)
		configFieldNames = make(map[reflect.Type]string)
		registered = registered[:0]
		mutators = make(map[reflect.Type]WrapperFunc)
	}

	t.Cleanup(clear)
	clear()

	for c, t := range cc {
		Register(c, t)
	}
}

// Registered all Configs that were passed to Register or RegisterDynamic. Each
// call will generate a new set of configs.
func Registered() []Config {
	res := make([]Config, 0, len(registered))
	for _, dyn := range registered {
		switch v := dyn.(type) {
		case Config:
			res = append(res, cloneDynamic(v).(Config))
		default:
			mut, ok := mutators[reflect.TypeOf(dyn)]
			if !ok || mut == nil {
				panic(fmt.Sprintf("Could not find transformer for dynamic integration %T", dyn))
			}
			res = append(res, mut(cloneDynamic(dyn)))
		}
	}
	return res
}

func cloneDynamic(in interface{}) interface{} {
	return reflect.New(reflect.TypeOf(in).Elem()).Interface()
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
		if wc, ok := c.(WrappedConfig); ok {
			data = wc.UnwrapConfig()
		}

		integrationType, ok := integrationTypes[reflect.TypeOf(data)]
		if !ok {
			panic(fmt.Sprintf("config not registered: %T", data))
		}

		switch integrationType {
		case TypeSingleton:
			field := outVal.FieldByName("XXX_Config_" + fieldName)
			field.Set(reflect.ValueOf(data))
		case TypeMultiplex, TypeEither:
			field := outVal.FieldByName("XXX_Configs_" + fieldName)
			field.Set(reflect.Append(field, reflect.ValueOf(data)))
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
				val := getConfig(field.Index(i).Interface())
				*configs = append(*configs, val)
			}
		default:
			val := getConfig(field.Interface())
			*configs = append(*configs, val)
		}
	}

	return nil
}

// getConfig returns a v2 config from an input.
func getConfig(in interface{}) Config {
	switch c := in.(type) {
	case Config:
		return c
	default:
		mut, ok := mutators[reflect.TypeOf(in)]
		if !ok {
			panic(fmt.Sprintf("unexpected type %T", c))
		}
		return mut(in)
	}
}

// getConfigTypeForIntegrations returns a dynamic struct type that has all of
// the same fields as out including the fields for the provided integrations.
//
// If marshal is true, interface{} will be used for integration types. This
// must be true when marshaling dynamic integrations, where their type will
// have changed.
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

	for _, dyn := range registered {
		// Fields use a prefix that's unlikely to collide with anything else.
		configTy := reflect.TypeOf(dyn)
		integrationType := integrationTypes[configTy]
		fieldName := configFieldNames[configTy]

		if integrationType == TypeSingleton || integrationType == TypeEither {
			fields = append(fields, reflect.StructField{
				Name: "XXX_Config_" + fieldName,
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s,omitempty"`, fieldName)),
				Type: configTy,
			})
		}
		if integrationType == TypeMultiplex || integrationType == TypeEither {
			fields = append(fields, reflect.StructField{
				Name: "XXX_Configs_" + fieldName,
				Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s_configs,omitempty"`, fieldName)),
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
