package v2

import (
	"context"
	"sync"
	"testing"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/go-kit/log"
)

// Test_controller_RunsIntegration ensures that integrations
// run.

func Test_controller_RunsIntegration(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	ctrl, err := NewController(
		util.TestLogger(t),
		newMockIntegrationConfigs(
			mockConfigForIntegration(t, funcIntegration(func(ctx context.Context) error {
				defer wg.Done()
				cancel()
				<-ctx.Done()
				return nil
			})),
		),
		shared.Globals{},
	)
	require.NoError(t, err, "failed to create controller")

	// Run the controller. The controller should immediately run our fake integration
	// which will cancel ctx and cause ctrl to exit.
	ctrl.Run(ctx)

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

		mockIntegration := funcIntegration(func(ctx context.Context) error {
			integrationsWg.Done()
			starts.Inc()
			<-ctx.Done()
			return nil
		})

		cfg := newMockIntegrationConfigs(
			&mockConfig{
				NameFunc:          func() string { return mockIntegrationName },
				ConfigEqualsFunc:  func(Config) bool { return !changed },
				ApplyDefaultsFunc: func(g shared.Globals) error { return nil },
				IdentifierFunc: func(shared.Globals) (string, error) {
					return mockIntegrationName, nil
				},
				NewIntegrationFunc: func(log.Logger, shared.Globals) (Integration, error) {
					integrationsWg.Add(1)
					return mockIntegration, nil
				},
			},
		)

		globals := shared.Globals{}
		ctrl, err := NewController(util.TestLogger(t), cfg, globals)
		require.NoError(t, err, "failed to create controller")

		sc := NewSyncController(t, ctrl)
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

type SyncController struct {
	inner   *Controller
	applyWg sync.WaitGroup

	stop     context.CancelFunc
	exitedCh chan struct{}
}

// NewSyncController makes calls to Controller synchronous. NewSyncController
// will start running the inner Controller and wait for it to update.
func NewSyncController(t *testing.T, inner *Controller) *SyncController {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	sc := &SyncController{
		inner:    inner,
		stop:     cancel,
		exitedCh: make(chan struct{}),
	}
	inner.OnUpdateDone = sc.applyWg.Done // Inform WG whenever an apply finishes

	// There's always immediately ony applied queued from any successfully created controller.
	sc.applyWg.Add(1)

	go func() {
		inner.Run(ctx)
		close(sc.exitedCh)
	}()

	sc.applyWg.Wait()
	return sc
}

func (sc *SyncController) UpdateController(c []Config, g shared.Globals) error {
	sc.applyWg.Add(1)

	if err := sc.inner.UpdateController(c, g); err != nil {
		sc.applyWg.Done() // The wg won't ever be finished now
		return err
	}

	sc.applyWg.Wait()
	return nil
}

func (sc *SyncController) Stop() {
	sc.stop()
	<-sc.exitedCh
}

const mockIntegrationName = "mock"

var (
	_ MetricsIntegration = (*mockConfig)(nil)
)

type mockConfig struct {
	NameFunc           func() string
	ApplyDefaultsFunc  func(shared.Globals) error
	ConfigEqualsFunc   func(Config) bool
	IdentifierFunc     func(shared.Globals) (string, error)
	NewIntegrationFunc func(log.Logger, shared.Globals) (Integration, error)
}

func (mc *mockConfig) RunIntegration(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (mc *mockConfig) Targets(ep Endpoint) []*targetgroup.Group {
	//TODO implement me
	panic("implement me")
}

func (mc *mockConfig) ScrapeConfigs(configs discovery.Configs) []*autoscrape.ScrapeConfig {
	//TODO implement me
	panic("implement me")
}

func (mc *mockConfig) InstanceKey(agentKey string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (mc *mockConfig) Name() string {
	return mc.NameFunc()
}

func (mc *mockConfig) ConfigEquals(c Config) bool {
	if mc.ConfigEqualsFunc != nil {
		return mc.ConfigEqualsFunc(c)
	}
	return false
}

func (mc *mockConfig) ApplyDefaults(g shared.Globals) error {
	return mc.ApplyDefaultsFunc(g)
}

func (mc *mockConfig) Identifier(g shared.Globals) (string, error) {
	return mc.IdentifierFunc(g)
}

func (mc *mockConfig) NewIntegration(l log.Logger, g shared.Globals) (Integration, error) {
	return mc.NewIntegrationFunc(l, g)
}

func (mc *mockConfig) WithNewIntegrationFunc(f func(log.Logger, shared.Globals) (Integration, error)) *mockConfig {
	return &mockConfig{
		NameFunc:           mc.NameFunc,
		ApplyDefaultsFunc:  mc.ApplyDefaultsFunc,
		ConfigEqualsFunc:   mc.ConfigEqualsFunc,
		IdentifierFunc:     mc.IdentifierFunc,
		NewIntegrationFunc: f,
	}
}

func mockConfigNameTuple(t *testing.T, name, id string) *mockConfig {
	t.Helper()

	return &mockConfig{
		NameFunc:          func() string { return name },
		IdentifierFunc:    func(_ shared.Globals) (string, error) { return id, nil },
		ApplyDefaultsFunc: func(g shared.Globals) error { return nil },
		NewIntegrationFunc: func(log.Logger, shared.Globals) (Integration, error) {
			return NoOpIntegration, nil
		},
	}
}

// mockConfigForIntegration returns a Config that will always return i.
func mockConfigForIntegration(t *testing.T, i Integration) *mockConfig {
	t.Helper()

	return &mockConfig{
		NameFunc:          func() string { return mockIntegrationName },
		ApplyDefaultsFunc: func(g shared.Globals) error { return nil },
		IdentifierFunc: func(shared.Globals) (string, error) {
			return mockIntegrationName, nil
		},
		NewIntegrationFunc: func(log.Logger, shared.Globals) (Integration, error) {
			return i, nil
		},
	}
}
