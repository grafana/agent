package integrations

import (
	"context"
	"sync"
	"testing"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

//
// Tests for Controller's utilization of the UpdateIntegration interface.
//

// TestController_UpdateIntegration ensures that the controller will call
// UpdateIntegration for integrations that support it.
func TestController_UpdateIntegration(t *testing.T) {
	var (
		integrationStartWg sync.WaitGroup
		applies, starts    atomic.Uint64
	)

	mockIntegration := mockUpdateIntegration{
		Integration: FuncIntegration(func(ctx context.Context) error {
			starts.Inc()
			integrationStartWg.Done()
			<-ctx.Done()
			return nil
		}),
		ApplyConfigFunc: func(Config) error {
			applies.Inc()
			return nil
		},
	}

	cfg := ControllerConfig{
		mockConfig{
			NameFunc:         func() string { return "mock" },
			ConfigEqualsFunc: func(Config) bool { return false },
			IdentifierFunc: func(IntegrationOptions) (string, error) {
				return "mock", nil
			},
			NewIntegrationFunc: func(IntegrationOptions) (Integration, error) {
				integrationStartWg.Add(1)
				return mockIntegration, nil
			},
		},
	}

	ctrl, err := NewController(cfg, IntegrationOptions{Logger: util.TestLogger(t)})
	require.NoError(t, err, "failed to create controller")

	sc := newSyncController(t, ctrl)

	// Wait for our integration to start.
	integrationStartWg.Wait()

	// Try to apply again.
	require.NoError(t, sc.ApplyConfig(cfg), "failed to re-apply config")
	integrationStartWg.Wait()

	sc.Stop()

	require.Equal(t, uint64(1), applies.Load(), "dynamic reload should have occured")
	require.Equal(t, uint64(1), starts.Load(), "restart should not have occured")
}

// TestController_UpdateIntegration ensures that the controller will remove
// integrations after an Update disables it.
func TestController_UpdateIntegration_Disabled(t *testing.T) {
	var (
		startWg, stopWg sync.WaitGroup
	)

	mockIntegration := mockUpdateIntegration{
		Integration: FuncIntegration(func(ctx context.Context) error {
			startWg.Done()
			defer stopWg.Done()
			<-ctx.Done()
			return nil
		}),
		ApplyConfigFunc: func(Config) error {
			return ErrDisabled
		},
	}

	cfg := ControllerConfig{
		mockConfig{
			NameFunc:         func() string { return "mock" },
			ConfigEqualsFunc: func(Config) bool { return false },
			IdentifierFunc: func(IntegrationOptions) (string, error) {
				return "mock", nil
			},
			NewIntegrationFunc: func(IntegrationOptions) (Integration, error) {
				startWg.Add(1)
				stopWg.Add(1)
				return mockIntegration, nil
			},
		},
	}

	ctrl, err := NewController(cfg, IntegrationOptions{Logger: util.TestLogger(t)})
	require.NoError(t, err, "failed to create controller")

	sc := newSyncController(t, ctrl)

	// Wait for our integration to start.
	startWg.Wait()

	// Try to apply again. This should pick up the ErrDisabled on apply and force
	// our itnegration to stop.
	require.NoError(t, sc.ApplyConfig(cfg), "failed to re-apply config")
	stopWg.Wait()

	sc.Stop()
}

type mockUpdateIntegration struct {
	Integration
	ApplyConfigFunc func(c Config) error
}

func (m mockUpdateIntegration) ApplyConfig(c Config) error {
	return m.ApplyConfigFunc(c)
}
