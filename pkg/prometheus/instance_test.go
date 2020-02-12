package prometheus

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

func TestInstanceConfig_ApplyDefaults(t *testing.T) {
	global := config.DefaultGlobalConfig
	cfg := &InstanceConfig{
		Name: "instance",
		ScrapeConfigs: []*config.ScrapeConfig{{
			JobName: "scrape",
			ServiceDiscoveryConfig: sd_config.ServiceDiscoveryConfig{
				StaticConfigs: []*targetgroup.Group{{
					Targets: []model.LabelSet{{
						model.AddressLabel: model.LabelValue("127.0.0.1:12345"),
					}},
					Labels: model.LabelSet{"cluster": "localhost"},
				}},
			},
		}},
	}

	cfg.ApplyDefaults(&global)
	for _, sc := range cfg.ScrapeConfigs {
		require.Equal(t, sc.ScrapeInterval, global.ScrapeInterval)
		require.Equal(t, sc.ScrapeTimeout, global.ScrapeTimeout)
		require.Equal(t, sc.RelabelConfigs, DefaultRelabelConfigs)
	}
}

// TestInstance tests that discovery and scraping are working by using a mock
// instance of the WAL storage and testing that samples get written to it.
// This test touches most of Instance and is enough for a basic integration test.
func TestInstance(t *testing.T) {
	scrapeAddr, closeSrv := getTestServer(t)
	defer closeSrv()

	walDir, err := ioutil.TempDir(os.TempDir(), "wal")
	require.NoError(t, err)
	defer os.RemoveAll(walDir)

	globalConfig := getTestGlobalConfig(t)

	cfg := getTestInstanceConfig(t, &globalConfig, scrapeAddr)
	cfg.WALTruncateFrequency = time.Hour
	cfg.RemoteFlushDeadline = time.Hour

	mockStorage := mockWalStorage{
		series: make(map[uint64]int),
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	inst, err := newInstance(globalConfig, cfg, walDir, logger, &mockStorage)
	defer inst.Stop()

	// Wait until mockWalStorage has had a series added to it.
	test.Poll(t, 30*time.Second, true, func() interface{} {
		mockStorage.mut.Lock()
		defer mockStorage.mut.Unlock()
		return len(mockStorage.series) > 0
	})
}

func getTestServer(t *testing.T) (addr string, closeFunc func()) {
	t.Helper()

	reg := prometheus.NewRegistry()

	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_metric_total",
	})
	testCounter.Inc()
	reg.MustRegister(testCounter)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	httpSrv := httptest.NewServer(handler)
	return httpSrv.Listener.Addr().String(), httpSrv.Close
}

func getTestGlobalConfig(t *testing.T) config.GlobalConfig {
	t.Helper()

	return config.GlobalConfig{
		ScrapeInterval:     model.Duration(time.Millisecond * 50),
		ScrapeTimeout:      model.Duration(time.Millisecond * 100),
		EvaluationInterval: model.Duration(time.Hour),
	}
}

func getTestInstanceConfig(t *testing.T, global *config.GlobalConfig, scrapeAddr string) InstanceConfig {
	t.Helper()

	scrapeCfg := config.DefaultScrapeConfig
	scrapeCfg.JobName = "test"
	scrapeCfg.ScrapeInterval = global.ScrapeInterval
	scrapeCfg.ScrapeTimeout = global.ScrapeTimeout
	scrapeCfg.ServiceDiscoveryConfig = sd_config.ServiceDiscoveryConfig{
		StaticConfigs: []*targetgroup.Group{{
			Targets: []model.LabelSet{{
				model.AddressLabel: model.LabelValue(scrapeAddr),
			}},
			Labels: model.LabelSet{},
		}},
	}

	cfg := DefaultInstanceConfig
	cfg.Name = "test"
	cfg.ScrapeConfigs = []*config.ScrapeConfig{&scrapeCfg}

	return cfg
}

type mockWalStorage struct {
	storage.Queryable

	mut    sync.Mutex
	series map[uint64]int
}

func (s *mockWalStorage) StartTime() (int64, error)                  { return 0, nil }
func (s *mockWalStorage) WriteStalenessMarkers(f func() int64) error { return nil }
func (s *mockWalStorage) Close() error                               { return nil }
func (s *mockWalStorage) Truncate(mint int64) error                  { return nil }

func (s *mockWalStorage) Appender() (storage.Appender, error) {
	return &mockAppender{s: s}, nil
}

type mockAppender struct {
	s *mockWalStorage
}

// Add adds a new series and sets its written count to 1.
func (a *mockAppender) Add(l labels.Labels, t int64, v float64) (uint64, error) {
	a.s.mut.Lock()
	defer a.s.mut.Unlock()

	hash := l.Hash()
	a.s.series[hash] = 1
	return hash, nil
}

// AddFast increments the number of writes to an existing series.
func (a *mockAppender) AddFast(l labels.Labels, ref uint64, t int64, v float64) error {
	a.s.mut.Lock()
	defer a.s.mut.Unlock()
	_, ok := a.s.series[ref]
	if !ok {
		return storage.ErrNotFound
	}

	a.s.series[ref]++
	return nil
}

func (a *mockAppender) Commit() error {
	return nil
}

func (a *mockAppender) Rollback() error {
	return nil
}
