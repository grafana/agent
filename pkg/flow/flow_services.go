package flow

import "github.com/grafana/agent/pkg/flow/internal/controller"

// GetServiceConsumers implements [service.Host]. It returns a slice of
// [component.Component] and [service.Service]s which declared a dependency on
// the named service.
func (f *Flow) GetServiceConsumers(serviceName string) []any {
	graph := f.loader.OriginalGraph()

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
			if svc := consumer.Service(); svc != nil {
				consumers = append(consumers, svc)
			}
		}
	}

	return consumers
}
