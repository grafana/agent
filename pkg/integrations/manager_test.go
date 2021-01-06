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
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"
)

// Test that embedded integration fields in the struct can be unmarshaled and
// remarshaled back out to text.
func TestConfig_Remarshal(t *testing.T) {
	RegisterIntegration(&testIntegrationA{})

	cfgText := `
scrape_integrations: true
replace_instance_label: true
integration_restart_backoff: 5s
use_hostname_label: true
test:
  text: Hello, world!
  truth: true
`
	var (
		cfg        ManagerConfig
		listenPort int = 12345
		listenHost string = "127.0.0.1"
	)
	require.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	// Listen port must be set before applying defaults. Normally applied by the
	// config package.
	cfg.ListenPort = &listenPort
	cfg.ListenHost = &listenHost

	outBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	fmt.Println(string(outBytes))
	require.YAMLEq(t, cfgText, string(outBytes))
}

func TestConfig_AddressRelabels(t *testing.T) {
	cfgText := `
agent: 
  enabled: true
`

	var (
		cfg        ManagerConfig
		listenPort int = 12345
		listenHost string = "127.0.0.1"
	)
	require.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	// Listen port must be set before applying defaults. Normally applied by the
	// config package.
	cfg.ListenPort = &listenPort
	cfg.ListenHost = &listenHost

	relabels, err := cfg.DefaultRelabelConfigs()
	require.NoError(t, err)

	// Ensure that the relabel configs are functional
	require.Len(t, relabels, 1)
	result := relabel.Process(labels.FromStrings("__address__", "127.0.0.1"), relabels...)

	expectHostname, _ := instance.Hostname()
	require.Equal(t, result.Get("instance"), expectHostname+":12345")
}

// TestManager_ValidInstanceConfigs ensures that the instance configs
// applied to the instance manager are valid.
func TestManager_ValidInstanceConfigs(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{integration: mock}

	integrations := map[Config]Integration{icfg: mock}
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

func TestManager_instanceConfigForIntegration(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{integration: mock}

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory, func(c *instance.Config) error {
		globalConfig := prom_config.DefaultConfig.GlobalConfig
		return c.ApplyDefaults(&globalConfig)
	})
	m, err := newManager(mockManagerConfig(), log.NewNopLogger(), im, nil)
	require.NoError(t, err)
	defer m.Stop()

	cfg := m.instanceConfigForIntegration(icfg, mock)

	// Validate that the generated MetricsPath is a valid URL path
	require.Len(t, cfg.ScrapeConfigs, 1)
	require.Equal(t, "/integrations/mock/metrics", cfg.ScrapeConfigs[0].MetricsPath)
}

// TestManager_NoIntegrationsScrape ensures that configs don't get generates
// when the ScrapeIntegrations flag is disabled.
func TestManager_NoIntegrationsScrape(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{integration: mock}

	integrations := map[Config]Integration{icfg: mock}
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
	icfg := mockConfig{integration: mock}

	noScrape := false
	mock.commonCfg.ScrapeIntegration = &noScrape

	integrations := map[Config]Integration{icfg: mock}
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
	icfg := mockConfig{integration: mock}

	integrations := map[Config]Integration{icfg: mock}

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
	icfg := mockConfig{integration: mock}

	integrations := map[Config]Integration{icfg: mock}
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
	icfg := mockConfig{integration: mock}

	integrations := map[Config]Integration{icfg: mock}
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

type mockConfig struct {
	integration *mockIntegration
}

func (c mockConfig) Name() string                { return "mock" }
func (c mockConfig) CommonConfig() config.Common { return c.integration.commonCfg }
func (c mockConfig) NewIntegration(_ log.Logger) (Integration, error) {
	return c.integration, nil
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

func mockManagerConfig() ManagerConfig {
	listenPort := 0
	listenHost := "127.0.0.1"
	return ManagerConfig{
		ScrapeIntegrations:        true,
		IntegrationRestartBackoff: 0,
		ListenPort:                &listenPort,
		ListenHost:                &listenHost,
	}
}
