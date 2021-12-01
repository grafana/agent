package integrations

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

//
// Tests for Controller's utilization of the core Integration interface.
//

// TestController_UniqueIdentifer ensures that integrations must not share a (name, id) tuple.
func TestController_UniqueIdentifier(t *testing.T) {
	controllerFromConfigs := func(t *testing.T, cc []Config) (*Controller, error) {
		t.Helper()
		return NewController(
			ControllerConfig(cc),
			IntegrationOptions{Logger: util.TestLogger(t)},
		)
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

// TestController_RunsIntegration ensures that integrations
// run.
func TestController_RunsIntegration(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	ctrl, err := NewController(
		ControllerConfig{
			mockConfigForIntegration(t, FuncIntegration(func(ctx context.Context) error {
				defer wg.Done()
				cancel()
				<-ctx.Done()
				return nil
			})),
		},
		IntegrationOptions{Logger: util.TestLogger(t)},
	)
	require.NoError(t, err, "failed to create controller")

	// Run the controller. The controller should immediately run our fake integration
	// which will cancel ctx and cause ctrl to exit.
	ctrl.Run(ctx)

	// Make sure that our integration exited too.
	wg.Wait()
}

// TestController_ConfigChanges ensures that integrations only get restarted
// when configs are no longer equal.
func TestController_ConfigChanges(t *testing.T) {
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

		cfg := ControllerConfig{
			mockConfig{
				NameFunc:         func() string { return "mock" },
				ConfigEqualsFunc: func(Config) bool { return !changed },
				IdentifierFunc: func(IntegrationOptions) (string, error) {
					return "mock", nil
				},
				NewIntegrationFunc: func(IntegrationOptions) (Integration, error) {
					integrationsWg.Add(1)
					return mockIntegration, nil
				},
			},
		}

		iopts := IntegrationOptions{Logger: util.TestLogger(t)}
		ctrl, err := NewController(cfg, iopts)
		require.NoError(t, err, "failed to create controller")

		sc := newSyncController(t, ctrl)
		require.NoError(t, sc.UpdateController(cfg, iopts), "failed to re-apply config")

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
	inner   *Controller
	applyWg sync.WaitGroup

	stop     context.CancelFunc
	exitedCh chan struct{}
}

// newSyncController makes calls to Controller synchronous. newSyncController
// will start running the inner controller.
func newSyncController(t *testing.T, inner *Controller) *syncController {
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

	// There's always immediately ony applied queued from any succesfully created controller.
	sc.applyWg.Add(1)

	go func() {
		err := inner.Run(ctx)
		require.NoError(t, err)
		close(sc.exitedCh)
	}()

	sc.applyWg.Wait()
	return sc
}

func (sc *syncController) UpdateController(c ControllerConfig, opts IntegrationOptions) error {
	sc.applyWg.Add(1)

	if err := sc.inner.UpdateController(c, opts); err != nil {
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

// TestController_IgnoresDisabledIntegration ensures that disabled integrations
// do not get run.
func TestController_IgnoredDisabledIntegration(t *testing.T) {
	cfg := ControllerConfig{
		mockConfig{
			NameFunc:         func() string { return "mock" },
			ConfigEqualsFunc: func(Config) bool { return false },
			IdentifierFunc: func(IntegrationOptions) (string, error) {
				return "mock", nil
			},
			NewIntegrationFunc: func(IntegrationOptions) (Integration, error) {
				return nil, fmt.Errorf("won't run integration: %w", ErrDisabled)
			},
		},
	}

	_, err := NewController(cfg, IntegrationOptions{Logger: util.TestLogger(t)})
	require.NoError(t, err, "error from NewIntegration should have been ignored")
}

type mockConfig struct {
	NameFunc           func() string
	ConfigEqualsFunc   func(Config) bool
	IdentifierFunc     func(IntegrationOptions) (string, error)
	NewIntegrationFunc func(IntegrationOptions) (Integration, error)
}

func (mc mockConfig) Name() string {
	return mc.NameFunc()
}

func (mc mockConfig) ConfigEquals(o Config) bool {
	if mc.ConfigEqualsFunc != nil {
		return mc.ConfigEqualsFunc(o)
	}
	return false
}

func (mc mockConfig) Identifier(o IntegrationOptions) (string, error) {
	return mc.IdentifierFunc(o)
}

func (mc mockConfig) NewIntegration(o IntegrationOptions) (Integration, error) {
	return mc.NewIntegrationFunc(o)
}

func mockConfigNameTuple(t *testing.T, name, id string) mockConfig {
	t.Helper()

	return mockConfig{
		NameFunc:       func() string { return name },
		IdentifierFunc: func(_ IntegrationOptions) (string, error) { return id, nil },
		NewIntegrationFunc: func(_ IntegrationOptions) (Integration, error) {
			return NoOpIntegration, nil
		},
	}
}

// mockConfigForIntegration returns a Config that will always return i.
func mockConfigForIntegration(t *testing.T, i Integration) mockConfig {
	t.Helper()

	return mockConfig{
		NameFunc: func() string { return "mock" },
		IdentifierFunc: func(io IntegrationOptions) (string, error) {
			return "mock", nil
		},
		NewIntegrationFunc: func(io IntegrationOptions) (Integration, error) {
			return i, nil
		},
	}
}
