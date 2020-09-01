package integrations

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

// TestManager_ValidInstanceConfigs ensures that the instance configs
// applied to the instance manager are valid.
func TestManager_ValidInstanceConfigs(t *testing.T) {
	mock := newMockIntegration()

	integrations := []Integration{mock}
	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, func(c *instance.Config) error {
		globalConfig := prom_config.DefaultConfig.GlobalConfig
		return c.ApplyDefaults(&globalConfig)
	})
	m, err := newManager(mockManagerConfig(), log.NewNopLogger(), im, integrations)
	require.NoError(t, err)
	defer m.Stop()

	// If the config doesn't show up in ListConfigs, it wasn't valid.
	test.Poll(t, time.Second, 1, func() interface{} {
		return len(im.ListConfigs())
	})
}

// TestManager_NoIntegrationsScrape ensures that configs don't get generates
// when the ScrapeIntegrations flag is disabled.
func TestManager_NoIntegrationsScrape(t *testing.T) {
	mock := newMockIntegration()

	integrations := []Integration{mock}
	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, func(c *instance.Config) error {
		globalConfig := prom_config.DefaultConfig.GlobalConfig
		return c.ApplyDefaults(&globalConfig)
	})

	cfg := mockManagerConfig()
	cfg.ScrapeIntegrations = false

	m, err := newManager(cfg, log.NewNopLogger(), im, integrations)
	require.NoError(t, err)
	defer m.Stop()

	// Normally we'd use test.Poll here, but since im.ListConfigs starts out with a
	// length of zero, test.Poll would immediately pass. Instead we want to wait for a
	// bit to make sure that the length of ListConfigs doesn't become non-zero.
	time.Sleep(time.Second)
	require.Zero(t, len(im.ListConfigs()))
}

// TestManager_NoIntegrationScrape ensures that configs don't get generates
// when the ScrapeIntegration flag is disabled on the integration.
func TestManager_NoIntegrationScrape(t *testing.T) {
	mock := newMockIntegration()

	noScrape := false
	mock.commonCfg.ScrapeIntegration = &noScrape

	integrations := []Integration{mock}
	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, func(c *instance.Config) error {
		globalConfig := prom_config.DefaultConfig.GlobalConfig
		return c.ApplyDefaults(&globalConfig)
	})

	m, err := newManager(mockManagerConfig(), log.NewNopLogger(), im, integrations)
	require.NoError(t, err)
	defer m.Stop()

	time.Sleep(time.Second)
	require.Zero(t, len(im.ListConfigs()))
}

// TestManager_StartsIntegrations tests that, when given an integration to
// launch, TestManager applies a config and runs the integration.
func TestManager_StartsIntegrations(t *testing.T) {
	mock := newMockIntegration()

	integrations := []Integration{mock}

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, nil)
	m, err := newManager(mockManagerConfig(), log.NewNopLogger(), im, integrations)
	require.NoError(t, err)
	defer m.Stop()

	test.Poll(t, time.Second, 1, func() interface{} {
		return len(im.ListConfigs())
	})

	// Check that the instance was set to run
	test.Poll(t, time.Second, 1, func() interface{} {
		return int(mock.startedCount.Load())
	})
}

func TestManager_RestartsIntegrations(t *testing.T) {
	mock := newMockIntegration()

	integrations := []Integration{mock}
	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, nil)
	m, err := newManager(mockManagerConfig(), log.NewNopLogger(), im, integrations)
	require.NoError(t, err)
	defer m.Stop()

	mock.err <- fmt.Errorf("I can't believe this horrible error happened")

	test.Poll(t, time.Second, 2, func() interface{} {
		return int(mock.startedCount.Load())
	})
}

func TestManager_GracefulStop(t *testing.T) {
	mock := newMockIntegration()

	integrations := []Integration{mock}
	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, nil)
	m, err := newManager(mockManagerConfig(), log.NewNopLogger(), im, integrations)
	require.NoError(t, err)

	test.Poll(t, time.Second, 1, func() interface{} {
		return int(mock.startedCount.Load())
	})

	m.Stop()

	time.Sleep(500 * time.Millisecond)
	require.Equal(t, 1, int(mock.startedCount.Load()), "graceful shutdown should not have restarted the integration")

	test.Poll(t, time.Second, false, func() interface{} {
		return mock.running.Load()
	})
}

type mockIntegration struct {
	commonCfg    config.Common
	startedCount *atomic.Uint32
	running      *atomic.Bool
	err          chan error
}

func newMockIntegration() *mockIntegration {
	return &mockIntegration{
		running:      atomic.NewBool(true),
		startedCount: atomic.NewUint32(0),
		err:          make(chan error),
	}
}

func (i *mockIntegration) Name() string                { return "mock" }
func (i *mockIntegration) CommonConfig() config.Common { return i.commonCfg }
func (i *mockIntegration) RegisterRoutes(r *mux.Router) error {
	r.Handle("/metrics", promhttp.Handler())
	return nil
}
func (i *mockIntegration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     "mock",
		MetricsPath: "/metrics",
	}}
}
func (i *mockIntegration) Run(ctx context.Context) error {
	i.startedCount.Inc()
	i.running.Store(true)
	defer i.running.Store(false)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-i.err:
		return err
	}
}

func mockInstanceFactory(_ instance.Config) (instance.ManagedInstance, error) {
	return instance.NoOpInstance{}, nil
}

func mockManagerConfig() Config {
	listenPort := 0
	return Config{
		ScrapeIntegrations:        true,
		IntegrationRestartBackoff: 0,
		ListenPort:                &listenPort,
	}
}
