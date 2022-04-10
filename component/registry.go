package component

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/hashicorp/hcl/v2"
)

var (
	registered = map[string]registration{}
)

type registration struct {
	name    string
	builder builder
}

type builder interface {
	BuildComponent(*BuildContext, *hcl.Block) (HCL, error)
}

// Registration is used when registering a component, holding the component's
// name and builder. The name of the component must be a list of
// period-delimited valid identifiers, such as `remote.http`.
type Registration[Config any] struct {
	Name           string
	BuildComponent func(l log.Logger, c Config) (Component[Config], error)
}

// Register registers the definition of a component. Register will panic if the
// name is in use by another component.
func Register[Config any](r Registration[Config]) {
	if _, exist := registered[r.Name]; exist {
		panic(fmt.Sprintf("Component name %q already registered", r.Name))
	}

	// TODO(rfratto): validate names

	registered[r.Name] = registration{
		name:    r.Name,
		builder: newRawBuilder(r),
	}
}

// RegistrySchema returns an HCL schema from the registered objects.
func RegistrySchema() *hcl.BodySchema {
	var schema hcl.BodySchema

	usedBlockSchemas := make(map[string]struct{})

	for _, rc := range registered {
		nameParts := strings.Split(rc.name, ".")

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
