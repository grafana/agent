package integrations

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
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

// setRegisteredIntegrations clears the set of registered integrations and
// overrides it with cc. Used by tests.
func setRegisteredIntegrations(cc []Config) {
	registeredIntegrations = registeredIntegrations[:0]
	configFieldNames = make(map[reflect.Type]string)

	for _, c := range cc {
		RegisterIntegration(c)
	}
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
		cfgType    = getConfigTypeForIntegrations(registeredIntegrations, inType)
		cfgPointer = reflect.New(cfgType)
		cfgVal     = cfgPointer.Elem()
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
		if cfgType.Field(i).PkgPath != "" {
			continue // Field is unexported: ignore.
		}
		cfgVal.Field(i).Set(inVal.Field(i))
	}
	if configs == nil {
		return nil, fmt.Errorf("integrations: Configs field not found in type: %T", v)
	}

	for _, c := range configs {
		// TODO(rfratto): check to see if integration is a singleton only

		fieldName, ok := configFieldNames[reflect.TypeOf(c)]
		if !ok {
			return nil, fmt.Errorf("integrations: cannot marshal unregistered Config type: %T", c)
		}
		field := cfgVal.FieldByName("XXX_Configs_" + fieldName)
		field.Set(reflect.Append(field, reflect.ValueOf(c)))
	}

	return cfgPointer.Interface(), nil
}

// UnmarshalYAML helps implement yaml.Unmarshaller for structs that have a
// Configs field that should be inlined in the YAML string.
func UnmarshalYAML(out interface{}, unmarshal func(interface{}) error) error {
	return unmarshalIntegrationsWithList(registeredIntegrations, out, unmarshal)
}

// unmarshalIntegrationsWithList unmarshals to a subtype of out that has a
// field added for every integration in integrations. Code adapted from
// Prometheus:
//
//   https://github.com/prometheus/prometheus/blob/511511324adfc4f4178f064cc104c2deac3335de/discovery/registry.go#L111
func unmarshalIntegrationsWithList(integrations []Config, out interface{}, unmarshal func(interface{}) error) error {
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
func getConfigTypeForIntegrations(integrations []Config, out reflect.Type) reflect.Type {
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
		// Fields use a prefix that's unlikely to collide with anything else.
		fields = append(fields, reflect.StructField{
			Name: "XXX_Config_" + cfg.Name(),
			Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s,omitempty"`, cfg.Name())),
			Type: reflect.TypeOf(cfg),
		})
		fields = append(fields, reflect.StructField{
			Name: "XXX_Configs_" + cfg.Name(),
			Tag:  reflect.StructTag(fmt.Sprintf(`yaml:"%s_configs,omitempty"`, cfg.Name())),
			Type: reflect.SliceOf(reflect.TypeOf(cfg)),
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
