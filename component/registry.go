package component

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-kit/log"
	"github.com/hashicorp/hcl/v2"
)

var (
	registered = map[string]Registration{}
)

// Registration is used when registering a component, holding the component's
// name and builder. The name of the component must be a list of
// period-delimited valid identifiers, such as `remote.http`.
type Registration struct {
	Name           string
	Config         Config
	BuildComponent func(o Options, c Config) (Component, error)
}

// Options are provided to a Component when it is being constructed.
type Options struct {
	// ID of the component. Guaranteed to be globally unique across all
	// components.
	ComponentID string
	Logger      log.Logger

	// HTTP address (ip:port) of the running process.
	HTTPAddr string

	// OnStateChange be be invoked at any time by a component to queue
	// re-processing input for components which depend on the changed component.
	OnStateChange func()
}

// CloneConfig reutrns a new zero value of the registered config type.
func (r Registration) CloneConfig() Config {
	return reflect.New(reflect.TypeOf(r.Config)).Interface()
}

// Register registers the definition of a component. Register will panic if the
// name is in use by another component.
func Register(r Registration) {
	if _, exist := registered[r.Name]; exist {
		panic(fmt.Sprintf("Component name %q already registered", r.Name))
	}

	// TODO(rfratto): validate names

	registered[r.Name] = r
}

// Get looks up a registered component by name.
func Get(name string) (Registration, bool) {
	r, ok := registered[name]
	return r, ok
}

// RegistrySchema returns an HCL schema from the registered objects.
func RegistrySchema() *hcl.BodySchema {
	var schema hcl.BodySchema

	usedBlockSchemas := make(map[string]struct{})

	for _, rc := range registered {
		nameParts := strings.Split(rc.Name, ".")

		genericNameList := append([]string{nameParts[0]}, mapToLabels(nameParts[1:])...)
		genericName := strings.Join(genericNameList, ".")
		if _, defined := usedBlockSchemas[genericName]; defined {
			// This block was already added; skip
			continue
		}
		usedBlockSchemas[genericName] = struct{}{}

		schema.Blocks = append(schema.Blocks, hcl.BlockHeaderSchema{
			Type:       nameParts[0],
			LabelNames: mapToLabels(nameParts[1:]),
		})
	}

	return &schema
}

func mapToLabels(in []string) []string {
	switch len(in) {
	case 0:
		return []string{"name"}
	case 1:
		return []string{"kind", "name"}
	default:
		panic("Unexpected long component name")
	}
}
