package integrations

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

//
// Tests for Controller's utilization of the core Integration interface.
//

// Test_controller_UniqueIdentifer ensures that integrations must not share a (name, id) tuple.
func Test_controller_UniqueIdentifier(t *testing.T) {
	controllerFromConfigs := func(t *testing.T, cc []Config) (*controller, error) {
		t.Helper()
		return newController(util.TestLogger(t), controllerConfig(cc), Globals{})
	}

	t.Run("different name, identifier", func(t *testing.T) {
		_, err := controllerFromConfigs(t, []Config{
			mockConfigNameTuple(t, "foo", "bar"),
			mockConfigNameTuple(t, "fizz", "buzz"),
		})
		require.NoError(t, err)
	})

	t.Run("same name, different identifier", func(t *testing.T) {
		_, err := controllerFromConfigs(t, []Config{
			mockConfigNameTuple(t, "foo", "bar"),
			mockConfigNameTuple(t, "foo", "buzz"),
		})
		require.NoError(t, err)
	})

	t.Run("same name, same identifier", func(t *testing.T) {
		_, err := controllerFromConfigs(t, []Config{
			mockConfigNameTuple(t, "foo", "bar"),
			mockConfigNameTuple(t, "foo", "bar"),
		})
		require.Error(t, err, `multiple instance names "bar" in integration "foo"`)
	})
}

// Test_controller_RunsIntegration ensures that integrations
// run.
func Test_controller_RunsIntegration(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	ctrl, err := newController(
		util.TestLogger(t),
		controllerConfig{
			mockConfigForIntegration(t, FuncIntegration(func(ctx context.Context) error {
				defer wg.Done()
				cancel()
				<-ctx.Done()
				return nil
			})),
		},
		Globals{},
	)
	require.NoError(t, err, "failed to create controller")

	// Run the controller. The controller should immediately run our fake integration
	// which will cancel ctx and cause ctrl to exit.
	ctrl.run(ctx)

	// Make sure that our integration exited too.
	wg.Wait()
}

// Test_controller_ConfigChanges ensures that integrations only get restarted
// when configs are no longer equal.
func Test_controller_ConfigChanges(t *testing.T) {
	tc := func(t *testing.T, changed bool) (timesRan uint64) {
		t.Helper()

		var integrationsWg sync.WaitGroup
		var starts atomic.Uint64

		mockIntegration := FuncIntegration(func(ctx context.Context) error {
			integrationsWg.Done()
			starts.Inc()
			<-ctx.Done()
			return nil
		})

		cfg := controllerConfig{
			mockConfig{
				NameFunc:         func() string { return mockIntegrationName },
				ConfigEqualsFunc: func(Config) bool { return !changed },
				IdentifierFunc: func(Globals) (string, error) {
					return mockIntegrationName, nil
				},
				NewIntegrationFunc: func(log.Logger, Globals) (Integration, error) {
					integrationsWg.Add(1)
					return mockIntegration, nil
				},
			},
		}

		globals := Globals{}
		ctrl, err := newController(util.TestLogger(t), cfg, globals)
		require.NoError(t, err, "failed to create controller")

		sc := newSyncController(t, ctrl)
		require.NoError(t, sc.UpdateController(cfg, globals), "failed to re-apply config")

		// Wait for our integrations to have been started
		integrationsWg.Wait()

		sc.Stop()
		return starts.Load()
	}

	t.Run("Unchanged", func(t *testing.T) {
		starts := tc(t, false)
		require.Equal(t, uint64(1), starts, "integration should only have started exactly once")
	})

	t.Run("Changed", func(t *testing.T) {
		starts := tc(t, true)
		require.Equal(t, uint64(2), starts, "integration should have started exactly twice")
	})
}

type syncController struct {
	inner   *controller
	applyWg sync.WaitGroup

	stop     context.CancelFunc
	exitedCh chan struct{}
}

// newSyncController makes calls to Controller synchronous. newSyncController
// will start running the inner controller and wait for it to update.
func newSyncController(t *testing.T, inner *controller) *syncController {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	sc := &syncController{
		inner:    inner,
		stop:     cancel,
		exitedCh: make(chan struct{}),
	}
	inner.onUpdateDone = sc.applyWg.Done // Inform WG whenever an apply finishes

	// There's always immediately ony applied queued from any successfully created controller.
	sc.applyWg.Add(1)

	go func() {
		inner.run(ctx)
		close(sc.exitedCh)
	}()

	sc.applyWg.Wait()
	return sc
}

func (sc *syncController) UpdateController(c controllerConfig, g Globals) error {
	sc.applyWg.Add(1)

	if err := sc.inner.UpdateController(c, g); err != nil {
		sc.applyWg.Done() // The wg won't ever be finished now
		return err
	}

	sc.applyWg.Wait()
	return nil
}

func (sc *syncController) Stop() {
	sc.stop()
	<-sc.exitedCh
}

// Test_controller_IgnoresDisabledIntegration ensures that disabled integrations
// do not get run.
func Test_controller_IgnoredDisabledIntegration(t *testing.T) {
	cfg := controllerConfig{
		mockConfig{
			NameFunc:         func() string { return mockIntegrationName },
			ConfigEqualsFunc: func(Config) bool { return false },
			IdentifierFunc: func(Globals) (string, error) {
				return mockIntegrationName, nil
			},
			NewIntegrationFunc: func(log.Logger, Globals) (Integration, error) {
				return nil, fmt.Errorf("won't run integration: %w", ErrDisabled)
			},
		},
	}

	_, err := newController(util.TestLogger(t), cfg, Globals{})
	require.NoError(t, err, "error from NewIntegration should have been ignored")
}

const mockIntegrationName = "mock"

type mockConfig struct {
	NameFunc           func() string
	ConfigEqualsFunc   func(Config) bool
	IdentifierFunc     func(Globals) (string, error)
	NewIntegrationFunc func(log.Logger, Globals) (Integration, error)
}

func (mc mockConfig) Name() string {
	return mc.NameFunc()
}

func (mc mockConfig) ConfigEquals(c Config) bool {
	if mc.ConfigEqualsFunc != nil {
		return mc.ConfigEqualsFunc(c)
	}
	return false
}

func (mc mockConfig) Identifier(g Globals) (string, error) {
	return mc.IdentifierFunc(g)
}

func (mc mockConfig) NewIntegration(l log.Logger, g Globals) (Integration, error) {
	return mc.NewIntegrationFunc(l, g)
}

func (mc mockConfig) WithNewIntegrationFunc(f func(log.Logger, Globals) (Integration, error)) mockConfig {
	return mockConfig{
		NameFunc:           mc.NameFunc,
		ConfigEqualsFunc:   mc.ConfigEqualsFunc,
		IdentifierFunc:     mc.IdentifierFunc,
		NewIntegrationFunc: f,
	}
}

func mockConfigNameTuple(t *testing.T, name, id string) mockConfig {
	t.Helper()

	return mockConfig{
		NameFunc:       func() string { return name },
		IdentifierFunc: func(_ Globals) (string, error) { return id, nil },
		NewIntegrationFunc: func(log.Logger, Globals) (Integration, error) {
			return NoOpIntegration, nil
		},
	}
}

// mockConfigForIntegration returns a Config that will always return i.
func mockConfigForIntegration(t *testing.T, i Integration) mockConfig {
	t.Helper()

	return mockConfig{
		NameFunc: func() string { return mockIntegrationName },
		IdentifierFunc: func(Globals) (string, error) {
			return mockIntegrationName, nil
		},
		NewIntegrationFunc: func(log.Logger, Globals) (Integration, error) {
			return i, nil
		},
	}
}
