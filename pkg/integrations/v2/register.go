package integrations

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

// Configs is a list of integrations. Note that Configs does not implement
// yaml.Unmarshaler or yaml.Marshaler. Use the UnmarshalYAML or MarshalYAML
// methods to deal with integrations defined from YAML.
type Configs []Config
