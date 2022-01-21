package v2

import (
	"context"
	"sync"
	"testing"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

//
// Tests for controller's utilization of the UpdateIntegration interface.
//

// Test_controller_UpdateIntegration ensures that the controller will call
// UpdateIntegration for integrations that support it.
func Test_controller_UpdateIntegration(t *testing.T) {
	var (
		integrationStartWg sync.WaitGroup
		applies, starts    atomic.Uint64
	)

	mockIntegration := mockUpdateIntegration{
		Integration: funcIntegration(func(ctx context.Context) error {
			starts.Inc()
			integrationStartWg.Done()
			<-ctx.Done()
			return nil
		}),
		ApplyConfigFunc: func(Config, shared.Globals) error {
			applies.Inc()
			return nil
		},
	}

	cfg := newMockIntegrationConfigs(
		&mockConfig{
			NameFunc:          func() string { return mockIntegrationName },
			ConfigEqualsFunc:  func(Config) bool { return false },
			ApplyDefaultsFunc: func(g shared.Globals) error { return nil },
			IdentifierFunc: func(shared.Globals) (string, error) {
				return mockIntegrationName, nil
			},
			NewIntegrationFunc: func(log.Logger, shared.Globals) (Integration, error) {
				integrationStartWg.Add(1)
				return mockIntegration, nil
			},
		},
	)

	ctrl, err := NewController(util.TestLogger(t), cfg, shared.Globals{})
	require.NoError(t, err, "failed to create controller")

	sc := NewSyncController(t, ctrl)

	// Wait for our integration to start.
	integrationStartWg.Wait()

	// Try to apply again.
	require.NoError(t, sc.UpdateController(cfg, ctrl.globals), "failed to re-apply config")
	integrationStartWg.Wait()

	sc.Stop()

	require.Equal(t, uint64(1), applies.Load(), "dynamic reload should have occurred")
	require.Equal(t, uint64(1), starts.Load(), "restart should not have occurred")
}

type mockUpdateIntegration struct {
	shared.Integration
	ApplyConfigFunc func(Config, shared.Globals) error
}

func (m mockUpdateIntegration) RunIntegration(ctx context.Context) error {
	return m.Run(ctx)
}

func (m mockUpdateIntegration) ApplyConfig(c Config, g shared.Globals) error {
	return m.ApplyConfigFunc(c, g)
}
