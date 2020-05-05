package instance

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
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

func TestConfig_ApplyDefaults(t *testing.T) {
	global := config.DefaultGlobalConfig
	cfg := &Config{
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

func TestInstance_Path(t *testing.T) {
	scrapeAddr, closeSrv := getTestServer(t)
	defer closeSrv()

	walDir, err := ioutil.TempDir(os.TempDir(), "wal")
	require.NoError(t, err)
	defer os.RemoveAll(walDir)

	globalConfig := getTestGlobalConfig(t)

	cfg := getTestConfig(t, &globalConfig, scrapeAddr)
	cfg.WALTruncateFrequency = time.Hour
	cfg.RemoteFlushDeadline = time.Hour

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	inst, err := New(globalConfig, cfg, walDir, logger)
	require.NoError(t, err)
	defer inst.Stop()

	// <walDir>/<inst.name> path should exist for WAL
	test.Poll(t, time.Second*5, true, func() interface{} {
		_, err := os.Stat(path.Join(walDir, "test"))
		return err == nil
	})
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

	cfg := getTestConfig(t, &globalConfig, scrapeAddr)
	cfg.WALTruncateFrequency = time.Hour
	cfg.RemoteFlushDeadline = time.Hour

	mockStorage := mockWalStorage{
		series: make(map[uint64]int),
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	inst, err := newInstance(globalConfig, cfg, nil, walDir, logger, &mockStorage)
	require.NoError(t, err)
	defer inst.Stop()

	// Wait until mockWalStorage has had a series added to it.
	test.Poll(t, 30*time.Second, true, func() interface{} {
		mockStorage.mut.Lock()
		defer mockStorage.mut.Unlock()
		return len(mockStorage.series) > 0
	})
}

// TestInstance_Recreate ensures that creating an instance with the same name twice
// does not cause any duplicate metrics registration that leads to a panic.
func TestInstance_Recreate(t *testing.T) {
	scrapeAddr, closeSrv := getTestServer(t)
	defer closeSrv()

	walDir, err := ioutil.TempDir(os.TempDir(), "wal")
	require.NoError(t, err)
	defer os.RemoveAll(walDir)

	globalConfig := getTestGlobalConfig(t)

	cfg := getTestConfig(t, &globalConfig, scrapeAddr)
	cfg.Name = "recreate_test"
	cfg.WALTruncateFrequency = time.Hour
	cfg.RemoteFlushDeadline = time.Hour

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	inst, err := New(globalConfig, cfg, walDir, logger)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	inst.Stop()

	// Recreate the instance, no panic should happen.
	require.NotPanics(t, func() {
		inst, err := New(globalConfig, cfg, walDir, logger)
		require.NoError(t, err)
		defer inst.Stop()

		time.Sleep(1 * time.Second)
	})
}

func TestMetricValueCollector(t *testing.T) {
	r := prometheus.NewRegistry()
	vc := NewMetricValueCollector(r, "this_should_be_tracked")

	shouldTrack := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "this_should_be_tracked",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})

	shouldTrack.Set(12345)

	shouldNotTrack := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "this_should_not_be_tracked",
	})

	r.MustRegister(shouldTrack, shouldNotTrack)

	vals, err := vc.GetValues("foo", "bar")
	require.NoError(t, err)
	require.Equal(t, []float64{12345}, vals)
}

func TestRemoteWriteMetricInterceptor_AllValues(t *testing.T) {
	r := prometheus.NewRegistry()
	vc := NewMetricValueCollector(r, "track")

	valueA := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "this_should_be_tracked",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})
	valueA.Set(12345)

	valueB := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "track_this_too",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})
	valueB.Set(67890)

	shouldNotReturn := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "track_this_but_label_does_not_match",
		ConstLabels: prometheus.Labels{
			"foo": "nope",
		},
	})

	r.MustRegister(valueA, valueB, shouldNotReturn)

	vals, err := vc.GetValues("foo", "bar")
	require.NoError(t, err)
	require.Equal(t, []float64{12345, 67890}, vals)
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

func getTestConfig(t *testing.T, global *config.GlobalConfig, scrapeAddr string) Config {
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

	cfg := DefaultConfig
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

func (s *mockWalStorage) Appender() storage.Appender {
	return &mockAppender{s: s}
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
func (a *mockAppender) AddFast(ref uint64, t int64, v float64) error {
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
