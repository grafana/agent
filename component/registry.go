package component

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-kit/log"
	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
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

// ModuleController is a mechanism responsible for allowing components to create other components via modules.
type ModuleController interface {
	// NewModule creates a new, un-started Module with a given ID. Multiple calls to
	// NewModule must provide unique values for id. The empty string is a valid unique
	// value for id.
	//
	// If id is non-empty, it must be a valid River identifier, matching the
	// regex /[A-Za-z_][A-Za-z0-9_]/.
	NewModule(id string, export ExportFunc) (Module, error)
}

// Module is a controller for running components within a Module.
type Module interface {
	// LoadConfig parses River config and loads it into the Module.
	// LoadConfig can be called multiple times, and called prior to
	// [Module.Run].
	LoadConfig(config []byte, args map[string]any) error

	// Run starts the Module. No components within the Module
	// will be run until Run is called.
	//
	// Run blocks until the provided context is canceled. The ID of a module as defined in
	// ModuleController.NewModule will not be released until Run returns.
	Run(context.Context)
}

// ExportFunc is used for onExport of the Module
type ExportFunc func(exports map[string]any)

// Options are provided to a component when it is being constructed. Options
// are static for the lifetime of a component.
type Options struct {
	// ModuleController allows for the creation of modules.
	ModuleController ModuleController

	// ID of the component. Guaranteed to be globally unique across all running
	// components.
	ID string

	// Logger the component may use for logging. Logs emitted with the logger
	// always include the component ID as a field.
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

	// Registerer allows components to add their own metrics. The registerer will
	// come pre-wrapped with the component ID. It is not necessary for components
	// to unregister metrics on shutdown.
	Registerer prometheus.Registerer

	// Tracer allows components to record spans. The tracer will include an
	// attribute denoting the component ID.
	Tracer trace.TracerProvider

	// GetServiceData retrieves data for a service by calling
	// [service.Service.Data] for the specified service.
	//
	// GetServiceData will return an error if the service does not exist or was
	// not listed as a dependency with the registration of the component.
	//
	// The result of GetServiceData may be cached as the value will not change at
	// runtime.
	GetServiceData func(name string) (interface{}, error)
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

	// An example Arguments value that the registered component expects to
	// receive as input. Components should provide the zero value of their
	// Arguments type here.
	Args Arguments

	// An example Exports value that the registered component may emit as output.
	// A component which does not expose exports must leave this set to nil.
	Exports Exports

	// NeedsServices holds the set of service names which this component depends
	// on to run. If NeedsServices includes an invalid service name (either
	// because of a cyclic dependency or the named service doesn't exist),
	// components will fail to evaluate.
	//
	// Modules which are loaded by the registered component will only be able to
	// access services in this list.
	NeedsServices []string

	// Build should construct a new component from an initial Arguments and set
	// of options.
	Build func(opts Options, args Arguments) (Component, error)
}

// CloneArguments returns a new zero value of the registered Arguments type.
func (r Registration) CloneArguments() Arguments {
	return reflect.New(reflect.TypeOf(r.Args)).Interface()
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

// validatePrefixMatch validates that no component has a name that is solely a prefix of another.
//
// For example, this will return an error if both a "remote" and "remote.http"
// component are defined.
func validatePrefixMatch(check parsedName, against map[string]parsedName) error {
	// add a trailing dot to each component name, so that we are always matching
	// complete segments.
	name := check.String() + "."
	for _, other := range against {
		otherName := other.String() + "."
		// if either is a prefix of the other, we have ambiguous names.
		if strings.HasPrefix(otherName, name) || strings.HasPrefix(name, otherName) {
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
