package flow

import (
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/service"
)

// GetServiceConsumers implements [service.Host]. It returns a slice of
// [component.Component] and [service.Service]s which declared a dependency on
// the named service.
func (f *Flow) GetServiceConsumers(serviceName string) []service.Consumer {
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

func serviceConsumersForGraph(graph *dag.Graph, serviceName string, includePeerServices bool) []service.Consumer {
	serviceNode, _ := graph.GetByID(serviceName).(*controller.ServiceNode)
	if serviceNode == nil {
		return nil
	}
	dependants := graph.Dependants(serviceNode)

	consumers := make([]service.Consumer, 0, len(dependants))

	for _, consumer := range dependants {
		// Only return instances of component.Component and service.Service.
		switch consumer := consumer.(type) {
		case *controller.ServiceNode:
			if !includePeerServices {
				continue
			}

			if svc := consumer.Service(); svc != nil {
				consumers = append(consumers, service.Consumer{
					Type:  service.ConsumerTypeService,
					ID:    consumer.NodeID(),
					Value: svc,
				})
			}
		}
	}

	return consumers
}
