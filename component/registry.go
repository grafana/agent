package component

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-kit/log"
	"github.com/grafana/regexp"
	"github.com/hashicorp/hcl/v2"
)

// The parsedName of a component is the parts of its name ("remote.http") split
// by the "." delimiter.
type parsedName []string

// String re-joins the parsed name by the "." delimiter.
func (pn parsedName) String() string { return strings.Join(pn, ".") }

var (
	// Globally registered components
	registered = map[string]Registration{}
	// Parsed names for components
	parsedNames = map[string]parsedName{}
)

// Options are provided to a component when it is being constructed. Options
// are static for the lifetime of a component.
type Options struct {
	// ID of the component. Guaranteed to be globally unique across all running
	// components.
	ID string

	// Logger the component may use for logging. The component ID will always be
	// set as a field.
	Logger log.Logger

	// A path to a directory with this component may use for storage. The path is
	// guaranteed to be unique across all running components.
	//
	// The directory may not exist when the component is created; components
	// should create the directory if needed.
	DataPath string

	// OnStateChange may be invoked at any time by a component whose Export value
	// changes. The Flow controller then will queue re-processing components
	// which depend on the changed component.
	//
	// OnStateChange will panic if e does not match the Exports type registered
	// by the component; a component must use the same Exports type for its
	// lifetime.
	OnStateChange func(e Exports)
}

// Registration describes a single component.
type Registration struct {
	// Name of the component. Must be a list of period-delimited valid
	// identifiers, such as "remote.s3". Components sharing a prefix must have
	// the same number of identifiers; it is valid to register "remote.s3" and
	// "remote.http" but not "remote".
	//
	// Components may not have more than 2 identifiers.
	//
	// Each identifier must start with a valid ASCII letter, and be followed by
	// any number of underscores or alphanumeric ASCII characters.
	Name string

	// A singleton component only supports one instance of itself across the
	// whole process. Normally, multiple components of the same type may be
	// created.
	//
	// The fully-qualified name of a component is the combination of HCL block
	// name and all of its labels. Fully-qualified names must be unique across
	// the process. Components which are *NOT* singletons automatically support
	// user-supplied identifiers:
	//
	//     // Fully-qualified names: remote.s3.object-a, remote.s3.object-b
	//     remote "s3" "object-a" { ... }
	//     remote "s3" "object-b" { ... }
	//
	// This allows for multiple instances of the same component to be defined.
	// However, components registered as a singleton do not support user-supplied
	// identifiers:
	//
	//     node_exporter { ... }
	//
	// This prevents the user from defining multiple instances of node_exporter
	// with different fully-qualified names.
	Singleton bool

	// An example Config value that the registered component expects to receive
	// as input. Components should provide the zero value of their config type
	// here.
	Config Config

	// An example Exports value that the registered component may emit as output.
	// A component which does not expose exports must leave this set to nil.
	Exports Exports

	// Build should construct a new component from an initial Config and set of
	// options.
	Build func(o Options, c Config) (Component, error)
}

// CloneConfig returns a new zero value of the registered Config type.
func (r Registration) CloneConfig() Config {
	return reflect.New(reflect.TypeOf(r.Config)).Interface()
}

// Register registers a component. Register will panic if the name is in use by
// another component, if the name is invalid, or if the component name has a
// suffix length mismatch with an existing component.
func Register(r Registration) {
	if _, exist := registered[r.Name]; exist {
		panic(fmt.Sprintf("Component name %q already registered", r.Name))
	}

	parsed, err := parseComponentName(r.Name)
	if err != nil {
		panic(fmt.Sprintf("invalid component name %q: %s", r.Name, err))
	}
	if err := validatePrefixMatch(parsed, parsedNames); err != nil {
		panic(err)
	}

	registered[r.Name] = r
	parsedNames[r.Name] = parsed
}

var identifierRegex = regexp.MustCompile("^[A-Za-z][0-9A-Za-z_]*$")

// parseComponentName parses and validates name. "remote.http" will return
// []string{"remote", "http"}.
func parseComponentName(name string) (parsedName, error) {
	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("missing name")
	}

	if len(parts) > 2 {
		return nil, fmt.Errorf("component name may only have 1 or 2 identifiers, found %d", len(parts))
	}

	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("found empty identifier")
		}

		if !identifierRegex.MatchString(part) {
			return nil, fmt.Errorf("identifier %q is not valid", part)
		}
	}

	return parts, nil
}

// validatePrefixMatch validates that components that share a prefix have the
// same length of identifiers in their names.
//
// For example, this will return an error if both a "remote" and "remote.http"
// component are defined.
func validatePrefixMatch(check parsedName, against map[string]parsedName) error {
	for _, other := range against {
		if other[0] != check[0] {
			continue
		}

		if len(other) != len(check) {
			return fmt.Errorf("%q cannot be used because it is incompatible with %q", check, other)
		}
	}

	return nil
}

// Get finds a registered component by name.
func Get(name string) (Registration, bool) {
	r, ok := registered[name]
	return r, ok
}

// RegistrySchema returns an HCL body schema using all registered components.
func RegistrySchema() *hcl.BodySchema {
	var schema hcl.BodySchema

	usedBlockSchemas := make(map[string]struct{})

	for _, rc := range registered {
		parsed := parsedNames[rc.Name]
		if parsed == nil {
			// This will never happen when using the exposed API, but may creep up
			// from tests that only set registered but not parsedNames.
			panic(rc.Name + " is missing from parsedNames map")
		}

		// The generic name of a component uses the first identifier of the
		// component and mapping the rest of the identifiers to labels.
		labels := labelNames(parsed, rc.Singleton)
		genericNameList := append([]string{parsed[0]}, labels...)
		genericName := strings.Join(genericNameList, ".")
		if _, defined := usedBlockSchemas[genericName]; defined {
			// Ignore blocks that were already added. This will happen when
			// processing a component for "remote.http" and "remote.s3", since both
			// of them are injected into the schema as "remote.kind.name".
			continue
		}
		usedBlockSchemas[genericName] = struct{}{}

		schema.Blocks = append(schema.Blocks, hcl.BlockHeaderSchema{
			Type:       parsed[0],
			LabelNames: labels,
		})
	}

	return &schema
}

func labelNames(in parsedName, singleton bool) []string {
	if singleton {
		// Component does *NOT* support a user-supplied identifier.
		switch len(in) {
		case 1:
			return []string{}
		case 2:
			return []string{"kind"}
		default:
			panic("Unexpected component name " + in.String())
		}
	}

	// Component supports a user-supplied identifier.
	switch len(in) {
	case 1:
		return []string{"name"}
	case 2:
		return []string{"kind", "name"}
	default:
		panic("Unexpected component name " + in.String())
	}
}
