package integrations

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"
)

const mockIntegrationName = "integration/mock"

func noOpValidator(*instance.Config) error { return nil }

// TestConfig_MarshalEmptyIntegrations ensures that an empty set of integrations
// can be marshaled correctly.
func TestConfig_MarshalEmptyIntegrations(t *testing.T) {
	cfgText := `
scrape_integrations: true
replace_instance_label: true
integration_restart_backoff: 5s
use_hostname_label: true
`
	var (
		cfg        ManagerConfig
		listenPort int    = 12345
		listenHost string = "127.0.0.1"
	)
	require.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	// Listen port must be set before applying defaults. Normally applied by the
	// config package.
	cfg.ListenPort = listenPort
	cfg.ListenHost = listenHost

	outBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err, "Failed creating integration")
	fmt.Println(string(outBytes))
	require.YAMLEq(t, cfgText, string(outBytes))
}

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
		listenPort int    = 12345
		listenHost string = "127.0.0.1"
	)
	require.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	// Listen port must be set before applying defaults. Normally applied by the
	// config package.
	cfg.ListenPort = listenPort
	cfg.ListenHost = listenHost

	outBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err, "Failed creating integration")
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
		listenPort int    = 12345
		listenHost string = "127.0.0.1"
	)
	require.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	// Listen port must be set before applying defaults. Normally applied by the
	// config package.
	cfg.ListenPort = listenPort
	cfg.ListenHost = listenHost

	expectHostname, _ := instance.Hostname()
	relabels := cfg.DefaultRelabelConfigs(expectHostname)

	// Ensure that the relabel configs are functional
	require.Len(t, relabels, 1)
	result := relabel.Process(labels.FromStrings("__address__", "127.0.0.1"), relabels...)

	require.Equal(t, result.Get("instance"), expectHostname+":12345")
}

func TestManager_instanceConfigForIntegration(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{Integration: mock}

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)
	m, err := NewManager(mockManagerConfig(), log.NewNopLogger(), im, noOpValidator)
	require.NoError(t, err)
	defer m.Stop()

	cfg := m.instanceConfigForIntegration(icfg, mock, mockManagerConfig())

	// Validate that the generated MetricsPath is a valid URL path
	require.Len(t, cfg.ScrapeConfigs, 1)
	require.Equal(t, "/integrations/mock/metrics", cfg.ScrapeConfigs[0].MetricsPath)
}

// TestManager_NoIntegrationsScrape ensures that configs don't get generates
// when the ScrapeIntegrations flag is disabled.
func TestManager_NoIntegrationsScrape(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{Integration: mock}

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)

	cfg := mockManagerConfig()
	cfg.ScrapeIntegrations = false
	cfg.Integrations = append(cfg.Integrations, &icfg)

	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
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
	icfg := mockConfig{Integration: mock}

	noScrape := false
	mock.CommonCfg.ScrapeIntegration = &noScrape

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)

	cfg := mockManagerConfig()
	cfg.Integrations = append(cfg.Integrations, icfg)

	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
	require.NoError(t, err)
	defer m.Stop()

	time.Sleep(time.Second)
	require.Zero(t, len(im.ListConfigs()))
}

// TestManager_StartsIntegrations tests that, when given an integration to
// launch, TestManager applies a config and runs the integration.
func TestManager_StartsIntegrations(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{Integration: mock}

	cfg := mockManagerConfig()
	cfg.Integrations = append(cfg.Integrations, icfg)

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)
	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
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
	icfg := mockConfig{Integration: mock}

	cfg := mockManagerConfig()
	cfg.Integrations = append(cfg.Integrations, icfg)

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)
	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
	require.NoError(t, err)
	defer m.Stop()

	mock.err <- fmt.Errorf("I can't believe this horrible error happened")

	test.Poll(t, time.Second, 2, func() interface{} {
		return int(mock.startedCount.Load())
	})
}

func TestManager_GracefulStop(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{Integration: mock}

	cfg := mockManagerConfig()
	cfg.Integrations = append(cfg.Integrations, icfg)

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)
	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
	require.NoError(t, err)

	test.Poll(t, time.Second, 1, func() interface{} {
		return int(mock.startedCount.Load())
	})

	m.Stop()

	time.Sleep(500 * time.Millisecond)
	require.Equal(t, 1, int(mock.startedCount.Load()), "graceful shutdown should not have restarted the Integration")

	test.Poll(t, time.Second, false, func() interface{} {
		return mock.running.Load()
	})
}

func TestManager_IntegrationEnabledToDisabledReload(t *testing.T) {
	mock := newMockIntegration()
	icfg := mockConfig{Integration: mock}
	cfg := mockManagerConfig()
	cfg.Integrations = append(cfg.Integrations, icfg)

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)
	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
	require.NoError(t, err)

	// Test for Enabled -> Disabled
	_ = m.ApplyConfig(generateMockConfigWithEnabledFlag(false))
	require.Len(t, m.integrations, 0, "Integration was disabled so should be removed from map")
	_, err = m.im.GetInstance(mockIntegrationName)
	require.Error(t, err, "This mock should not exist")

	// test for Disabled -> Enabled
	_ = m.ApplyConfig(generateMockConfigWithEnabledFlag(true))
	require.Len(t, m.integrations, 1, "Integration was enabled so should be here")
	_, err = m.im.GetInstance(mockIntegrationName)
	require.NoError(t, err, "This mock should exist")
	require.Len(t, m.im.ListInstances(), 1, "This instance should exist")
}

func TestManager_IntegrationDisabledToEnabledReload(t *testing.T) {
	mock := newMockIntegration()
	mock.CommonCfg.Enabled = false
	icfg := mockConfig{Integration: mock}

	cfg := mockManagerConfig()
	cfg.Integrations = append(cfg.Integrations, icfg)

	im := instance.NewBasicManager(instance.DefaultBasicManagerConfig, log.NewNopLogger(), mockInstanceFactory)
	m, err := NewManager(cfg, log.NewNopLogger(), im, noOpValidator)
	require.NoError(t, err)
	require.Len(t, m.integrations, 0, "Integration was disabled so should be removed from map")
	_, err = m.im.GetInstance(mockIntegrationName)
	require.Error(t, err, "This mock should not exist")

	// test for Disabled -> Enabled

	_ = m.ApplyConfig(generateMockConfigWithEnabledFlag(true))
	require.Len(t, m.integrations, 1, "Integration was enabled so should be here")
	_, err = m.im.GetInstance(mockIntegrationName)
	require.NoError(t, err, "This mock should exist")
	require.Len(t, m.im.ListInstances(), 1, "This instance should exist")
}

func generateMockConfigWithEnabledFlag(enabled bool) ManagerConfig {
	enabledMock := newMockIntegration()
	enabledMock.CommonCfg.Enabled = enabled
	enabledConfig := mockConfig{Integration: enabledMock}
	enabledManagerConfig := mockManagerConfig()
	enabledManagerConfig.Integrations = append(enabledManagerConfig.Integrations, enabledConfig)
	return enabledManagerConfig
}

type mockConfig struct {
	Integration *mockIntegration `yaml:"mock"`
}

// Equal is used for cmp.Equal, since otherwise mockConfig can't be compared to itself.
func (c mockConfig) Equal(other mockConfig) bool { return c.Integration == other.Integration }

func (c mockConfig) Name() string                { return "mock" }
func (c mockConfig) CommonConfig() config.Common { return c.Integration.CommonCfg }
func (c mockConfig) NewIntegration(_ log.Logger) (Integration, error) {
	return c.Integration, nil
}

type mockIntegration struct {
	CommonCfg    config.Common `yaml:",inline"`
	startedCount *atomic.Uint32
	running      *atomic.Bool
	err          chan error
}

func newMockIntegration() *mockIntegration {
	return &mockIntegration{
		running:      atomic.NewBool(true),
		startedCount: atomic.NewUint32(0),
		err:          make(chan error),
		CommonCfg:    config.Common{Enabled: true},
	}
}

func (i *mockIntegration) MetricsHandler() (http.Handler, error) {
	return promhttp.Handler(), nil
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
		ListenPort:                listenPort,
		ListenHost:                listenHost,
	}
}
