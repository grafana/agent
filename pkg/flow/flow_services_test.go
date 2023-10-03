package flow

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/internal/testservices"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestServices(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		startedSvc = util.NewWaitTrigger()

		svc = &testservices.Fake{
			RunFunc: func(ctx context.Context, _ service.Host) error {
				startedSvc.Trigger()

				<-ctx.Done()
				return nil
			},
		}
	)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svc)

	ctrl := New(opts)
	require.NoError(t, ctrl.LoadSource(makeEmptyFile(t), nil))

	// Start the controller. This should cause our service to run.
	go ctrl.Run(ctx)

	require.NoError(t, startedSvc.Wait(5*time.Second), "Service did not start")
}

func TestServices_Configurable(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	type ServiceOptions struct {
		Name string `river:"name,attr"`
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		updateCalled = util.NewWaitTrigger()

		svc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{
					Name:       "fake",
					ConfigType: ServiceOptions{},
				}
			},

			UpdateFunc: func(newConfig any) error {
				defer updateCalled.Trigger()

				require.IsType(t, ServiceOptions{}, newConfig)
				require.Equal(t, "John Doe", newConfig.(ServiceOptions).Name)
				return nil
			},
		}
	)

	f, err := ParseSource(t.Name(), []byte(`
		fake {
			name = "John Doe"
		}
	`))
	require.NoError(t, err)
	require.NotNil(t, f)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svc)

	ctrl := New(opts)

	require.NoError(t, ctrl.LoadSource(f, nil))

	// Start the controller. This should cause our service to run.
	go ctrl.Run(ctx)

	require.NoError(t, updateCalled.Wait(5*time.Second), "Service was not configured")
}

// TestServices_Configurable_Optional ensures that a service with optional
// arguments is configured properly even when it is not defined in the config
// file.
func TestServices_Configurable_Optional(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	type ServiceOptions struct {
		Name string `river:"name,attr,optional"`
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		updateCalled = util.NewWaitTrigger()

		svc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{
					Name:       "fake",
					ConfigType: ServiceOptions{},
				}
			},

			UpdateFunc: func(newConfig any) error {
				defer updateCalled.Trigger()

				require.IsType(t, ServiceOptions{}, newConfig)
				require.Equal(t, ServiceOptions{}, newConfig.(ServiceOptions))
				return nil
			},
		}
	)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svc)

	ctrl := New(opts)

	require.NoError(t, ctrl.LoadSource(makeEmptyFile(t), nil))

	// Start the controller. This should cause our service to run.
	go ctrl.Run(ctx)

	require.NoError(t, updateCalled.Wait(5*time.Second), "Service was not configured")
}

func TestFlow_GetServiceConsumers(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	var (
		svcA = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{
					Name: "svc_a",
				}
			},
		}

		svcB = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{
					Name:      "svc_b",
					DependsOn: []string{"svc_a"},
				}
			},
		}
	)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svcA, svcB)

	ctrl := New(opts)
	defer cleanUpController(ctrl)
	require.NoError(t, ctrl.LoadSource(makeEmptyFile(t), nil))

	expectConsumers := []service.Consumer{{
		Type:  service.ConsumerTypeService,
		ID:    "svc_b",
		Value: svcB,
	}}
	require.Equal(t, expectConsumers, ctrl.GetServiceConsumers("svc_a"))
}

func TestFlow_GetServiceConsumers_Modules(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	componentBuilt := util.NewWaitTrigger()

	var (
		svc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{Name: "service"}
			},
		}

		registry = controller.RegistryMap{
			"module_loader": component.Registration{
				Name:          "module_loader",
				Args:          struct{}{},
				NeedsServices: []string{"service"},
				Build: func(opts component.Options, _ component.Arguments) (component.Component, error) {
					mod, err := opts.ModuleController.NewModule("", nil)
					require.NoError(t, err, "Failed to create module")

					err = mod.LoadConfig([]byte(`service_consumer "example" {}`), nil)
					require.NoError(t, err, "Failed to load module config")

					return &testcomponents.Fake{
						RunFunc: func(ctx context.Context) error {
							mod.Run(ctx)
							<-ctx.Done()
							return nil
						},
					}, nil
				},
			},

			"service_consumer": component.Registration{
				Name:          "service_consumer",
				Args:          struct{}{},
				NeedsServices: []string{"service"},
				Build: func(_ component.Options, _ component.Arguments) (component.Component, error) {
					componentBuilt.Trigger()
					return &testcomponents.Fake{}, nil
				},
			},
		}
	)

	cfg := `module_loader "example" {}`

	f, err := ParseSource(t.Name(), []byte(cfg))
	require.NoError(t, err)
	require.NotNil(t, f)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svc)

	ctrl := newController(controllerOptions{
		Options:           opts,
		ComponentRegistry: registry,
		ModuleRegistry:    newModuleRegistry(),
	})
	require.NoError(t, ctrl.LoadSource(f, nil))
	go ctrl.Run(ctx)

	require.NoError(t, componentBuilt.Wait(5*time.Second), "Component should have been built")

	consumers := ctrl.GetServiceConsumers("service")
	require.Len(t, consumers, 2, "There should be a consumer for the module loader and the module's component")
}

func TestComponents_Using_Services(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		componentBuilt = util.NewWaitTrigger()
		serviceStarted = util.NewWaitTrigger()

		serviceStartedCount = atomic.NewInt64(0)
	)

	var (
		dependencySvc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{Name: "dependency"}
			},

			RunFunc: func(ctx context.Context, host service.Host) error {
				if serviceStartedCount.Add(1) > 1 {
					require.FailNow(t, "service should only be started once by the root controller")
				}

				serviceStarted.Trigger()

				<-ctx.Done()
				return nil
			},
		}

		nonDependencySvc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{Name: "non_dependency"}
			},
		}

		registry = controller.RegistryMap{
			"service_consumer": component.Registration{
				Name:          "service_consumer",
				Args:          struct{}{},
				NeedsServices: []string{"dependency"},
				Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
					// Call Trigger in a defer so we can make some extra assertions before
					// the test exits.
					defer componentBuilt.Trigger()

					_, err := opts.GetServiceData("dependency")
					require.NoError(t, err, "component should be able to access services it depends on")

					_, err = opts.GetServiceData("non_dependency")
					require.Error(t, err, "component should not be able to access services it doesn't depend on")

					_, err = opts.GetServiceData("does_not_exist")
					require.Error(t, err, "component should not be able to access non-existent service")

					return &testcomponents.Fake{}, nil
				},
			},
		}
	)

	cfg := `
		service_consumer "example" {}
	`

	f, err := ParseSource(t.Name(), []byte(cfg))
	require.NoError(t, err)
	require.NotNil(t, f)

	opts := testOptions(t)
	opts.Services = append(opts.Services, dependencySvc, nonDependencySvc)

	ctrl := newController(controllerOptions{
		Options:           opts,
		ComponentRegistry: registry,
		ModuleRegistry:    newModuleRegistry(),
	})
	require.NoError(t, ctrl.LoadSource(f, nil))
	go ctrl.Run(ctx)

	require.NoError(t, componentBuilt.Wait(5*time.Second), "Component should have been built")
	require.NoError(t, serviceStarted.Wait(5*time.Second), "Service should have been started")
}

func TestComponents_Using_Services_In_Modules(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	componentBuilt := util.NewWaitTrigger()

	var (
		propagatedSvc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{Name: "propagated_service"}
			},
		}

		nonPropagatedSvc = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{Name: "non_propagated_service"}
			},
		}

		registry = controller.RegistryMap{
			"module_loader": component.Registration{
				Name:          "module_loader",
				Args:          struct{}{},
				NeedsServices: []string{"propagated_service"},
				Build: func(opts component.Options, _ component.Arguments) (component.Component, error) {
					mod, err := opts.ModuleController.NewModule("", nil)
					require.NoError(t, err, "Failed to create module")

					err = mod.LoadConfig([]byte(`service_consumer "example" {}`), nil)
					require.NoError(t, err, "Failed to load module config")

					return &testcomponents.Fake{
						RunFunc: func(ctx context.Context) error {
							mod.Run(ctx)
							<-ctx.Done()
							return nil
						},
					}, nil
				},
			},

			"service_consumer": component.Registration{
				Name:          "service_consumer",
				Args:          struct{}{},
				NeedsServices: []string{"propagated_service"},
				Build: func(opts component.Options, _ component.Arguments) (component.Component, error) {
					// Call Trigger in a defer so we can make some extra assertions before
					// the test exits.
					defer componentBuilt.Trigger()

					_, err := opts.GetServiceData("propagated_service")
					require.NoError(t, err, "component should be able to access services that were propagated to it")

					return &testcomponents.Fake{}, nil
				},
			},
		}
	)

	cfg := `module_loader "example" {}`

	f, err := ParseSource(t.Name(), []byte(cfg))
	require.NoError(t, err)
	require.NotNil(t, f)

	opts := testOptions(t)
	opts.Services = append(opts.Services, propagatedSvc, nonPropagatedSvc)

	ctrl := newController(controllerOptions{
		Options:           opts,
		ComponentRegistry: registry,
		ModuleRegistry:    newModuleRegistry(),
	})
	require.NoError(t, ctrl.LoadSource(f, nil))
	go ctrl.Run(ctx)

	require.NoError(t, componentBuilt.Wait(5*time.Second), "Component should have been built")
}

func makeEmptyFile(t *testing.T) *Source {
	t.Helper()

	f, err := ParseSource(t.Name(), nil)
	require.NoError(t, err)
	require.NotNil(t, f)

	return f
}
