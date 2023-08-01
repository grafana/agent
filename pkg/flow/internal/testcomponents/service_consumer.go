package testcomponents

import (
	"context"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/testservices"
)

func init() {
	component.Register(component.Registration{
		Name: "testcomponents.service_consumer",
		Args: ServiceConsumerArguments{},
		NeedsServices: []string{
			"hook", // From testservices.Hook
		},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewServiceConsumer(opts, args.(ServiceConsumerArguments))
		},
	})
}

// ServiceConsumerArguments describes how to configure the
// testcomponents.service_consumer component.
type ServiceConsumerArguments struct{}

// ServiceConsumer implements the testcomponents.service_consumer component.
type ServiceConsumer struct {
	opts component.Options
	log  log.Logger
}

// NewServiceConsumer creates a new service consumer component.
func NewServiceConsumer(o component.Options, args ServiceConsumerArguments) (*ServiceConsumer, error) {
	hookData, err := o.GetServiceData("hook")
	if err != nil {
		return nil, err
	}
	hooks := hookData.(testservices.Hooks)
	hooks.OnComponentCreate(o, args)

	sc := &ServiceConsumer{
		opts: o,
		log:  o.Logger,
	}
	if err := sc.Update(args); err != nil {
		return nil, err
	}
	return sc, nil
}

var (
	_ component.Component      = (*Passthrough)(nil)
	_ component.DebugComponent = (*Passthrough)(nil)
)

// Run implements Component.
func (sc *ServiceConsumer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements Component.
func (sc *ServiceConsumer) Update(args component.Arguments) error {
	// no-op
	return nil
}
