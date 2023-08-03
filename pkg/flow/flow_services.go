package flow

import (
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
)

// GetServiceConsumers implements [service.Host]. It returns a slice of
// [component.Component] and [service.Service]s which declared a dependency on
// the named service.
func (f *Flow) GetServiceConsumers(serviceName string) []any {
	consumers := serviceConsumersForGraph(f.loader.OriginalGraph(), serviceName, true)

	// Iterate through all modules to find other components that depend on the
	// service. Peer services aren't checked here, since the services are always
	// a subset of the services from the root controller.
	for _, mod := range f.modules.List() {
		moduleGraph := mod.f.loader.OriginalGraph()
		consumers = append(consumers, serviceConsumersForGraph(moduleGraph, serviceName, false)...)
	}

	return consumers
}

func serviceConsumersForGraph(graph *dag.Graph, serviceName string, includePeerServices bool) []any {
	serviceNode, _ := graph.GetByID(serviceName).(*controller.ServiceNode)
	if serviceNode == nil {
		return nil
	}
	dependants := graph.Dependants(serviceNode)

	consumers := make([]any, 0, len(dependants))

	for _, consumer := range dependants {
		// Only return instances of component.Component and service.Service.
		switch consumer := consumer.(type) {
		case *controller.ComponentNode:
			if c := consumer.Component(); c != nil {
				consumers = append(consumers, c)
			}

		case *controller.ServiceNode:
			if !includePeerServices {
				continue
			}

			if svc := consumer.Service(); svc != nil {
				consumers = append(consumers, svc)
			}
		}
	}

	return consumers
}
