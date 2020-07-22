// Package instance provides a mini Prometheus scraper and remote_writer.
package instance

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prom/wal"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/pkg/relabel"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"gopkg.in/yaml.v2"
)

var (
	remoteWriteMetricName = "queue_highest_sent_timestamp_seconds"
	managerMtx            sync.Mutex
)

// Default configuration values
var (
	// DefaultRelabelConfigs defines a list of relabel_configs that will
	// be automatically appended to the end of all Prometheus
	// configurations.
	DefaultRelabelConfigs = []*relabel.Config{
		// Add __host__ from Kubernetes node name
		{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_node_name"},
			TargetLabel:  "__host__",
			Action:       relabel.Replace,
			Separator:    ";",
			Regex:        relabel.MustNewRegexp("(.*)"),
			Replacement:  "$1",
		},
	}

	DefaultConfig = Config{
		HostFilter:           false,
		WALTruncateFrequency: 1 * time.Minute,
		RemoteFlushDeadline:  1 * time.Minute,
		WriteStaleOnShutdown: false,
	}
)

// Config is a specific agent that runs within the overall Prometheus
// agent. It has its own set of scrape_configs and remote_write rules.
type Config struct {
	Name          string                      `yaml:"name" json:"name"`
	HostFilter    bool                        `yaml:"host_filter" json:"host_filter"`
	ScrapeConfigs []*config.ScrapeConfig      `yaml:"scrape_configs,omitempty" json:"scrape_configs,omitempty"`
	RemoteWrite   []*config.RemoteWriteConfig `yaml:"remote_write,omitempty" json:"remote_write,omitempty"`

	// How frequently the WAL should be truncated.
	WALTruncateFrequency time.Duration `yaml:"wal_truncate_frequency,omitempty" json:"wal_truncate_frequency,omitempty"`

	RemoteFlushDeadline  time.Duration `yaml:"remote_flush_deadline,omitempty" json:"remote_flush_deadline,omitempty"`
	WriteStaleOnShutdown bool          `yaml:"write_stale_on_shutdown,omitempty" json:"write_stale_on_shutdown,omitempty"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c Config) MarshalYAML() (interface{}, error) {
	// We want users to be able to marshal instance.Configs directly without
	// *needing* to call instance.MarshalConfig, so we call it internally
	// here and return a map.
	bb, err := MarshalConfig(&c, false)
	if err != nil {
		return nil, err
	}

	// Use a yaml.MapSlice rather than a map[string]interface{} so
	// order of keys is retained compared to just calling MarshalConfig.
	var m yaml.MapSlice
	if err := yaml.Unmarshal(bb, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// ApplyDefaults applies default configurations to the configuration to all
// values that have not been changed to their non-zero value. ApplyDefaults
// also validates the config.
func (c *Config) ApplyDefaults(global *config.GlobalConfig) error {
	switch {
	case c.Name == "":
		return errors.New("missing instance name")
	case c.WALTruncateFrequency <= 0:
		return errors.New("wal_truncate_frequency must be greater than 0s")
	case c.RemoteFlushDeadline <= 0:
		return errors.New("remote_flush_deadline must be greater than 0s")
	}

	jobNames := map[string]struct{}{}
	for _, sc := range c.ScrapeConfigs {
		if sc == nil {
			return fmt.Errorf("empty or null scrape config section")
		}

		// First set the correct scrape interval, then check that the timeout
		// (inferred or explicit) is not greater than that.
		if sc.ScrapeInterval == 0 {
			sc.ScrapeInterval = global.ScrapeInterval
		}
		if sc.ScrapeTimeout > sc.ScrapeInterval {
			return fmt.Errorf("scrape timeout greater than scrape interval for scrape config with job name %q", sc.JobName)
		}
		if time.Duration(sc.ScrapeInterval) > c.WALTruncateFrequency {
			return fmt.Errorf("scrape interval greater than wal_truncate_frequency for scrape config with job name %q", sc.JobName)
		}
		if sc.ScrapeTimeout == 0 {
			if global.ScrapeTimeout > sc.ScrapeInterval {
				sc.ScrapeTimeout = sc.ScrapeInterval
			} else {
				sc.ScrapeTimeout = global.ScrapeTimeout
			}
		}

		if _, exists := jobNames[sc.JobName]; exists {
			return fmt.Errorf("found multiple scrape configs with job name %q", sc.JobName)
		}
		jobNames[sc.JobName] = struct{}{}

		sc.RelabelConfigs = append(sc.RelabelConfigs, DefaultRelabelConfigs...)
	}

	rwNames := map[string]struct{}{}
	for _, cfg := range c.RemoteWrite {
		if cfg == nil {
			return fmt.Errorf("empty or null remote write config section")
		}

		// Typically Prometheus ignores empty names here, but we need to assign a
		// unique name to the config so we can pull metrics from it when running
		// an instance.
		var generatedName bool
		if cfg.Name == "" {
			hash, err := getHash(cfg)
			if err != nil {
				return err
			}

			// We have to add the name of the instance to ensure that generated metrics
			// are unique across multiple agent instances. The remote write queues currently
			// globally register their metrics so we can't inject labels here.
			cfg.Name = c.Name + "-" + hash[:6]
			generatedName = true
		}

		if _, exists := rwNames[cfg.Name]; exists {
			if generatedName {
				return fmt.Errorf("found two identical remote_write configs")
			}
			return fmt.Errorf("found duplicate remote write configs with name %q", cfg.Name)
		}
		rwNames[cfg.Name] = struct{}{}
	}

	return nil
}

type walStorageFactory func(reg prometheus.Registerer) (walStorage, error)

// Instance is an individual metrics collector and remote_writer.
type Instance struct {
	cfg       Config
	globalCfg config.GlobalConfig
	logger    log.Logger

	reg    prometheus.Registerer
	newWal walStorageFactory

	vc *MetricValueCollector

	readyScrapeManager *scrape.ReadyScrapeManager
}

// New creates a new Instance with a directory for storing the WAL. The instance
// will not start until Run is called on the instance.
func New(globalCfg config.GlobalConfig, cfg Config, walDir string, logger log.Logger) (*Instance, error) {
	logger = log.With(logger, "instance", cfg.Name)

	instWALDir := filepath.Join(walDir, cfg.Name)

	reg := prometheus.WrapRegistererWith(prometheus.Labels{
		"instance_name": cfg.Name,
	}, prometheus.DefaultRegisterer)

	newWal := func(reg prometheus.Registerer) (walStorage, error) {
		return wal.NewStorage(logger, reg, instWALDir)
	}

	return newInstance(globalCfg, cfg, reg, logger, newWal)
}

func newInstance(globalCfg config.GlobalConfig, cfg Config, reg prometheus.Registerer, logger log.Logger, newWal walStorageFactory) (*Instance, error) {
	vc := NewMetricValueCollector(prometheus.DefaultGatherer, remoteWriteMetricName)

	i := &Instance{
		cfg:       cfg,
		globalCfg: globalCfg,
		logger:    logger,
		vc:        vc,

		reg:    reg,
		newWal: newWal,

		readyScrapeManager: &scrape.ReadyScrapeManager{},
	}

	return i, nil
}

// Run starts the instance and will run until an error happens during running
// or until context was canceled.
//
// Run can be re-called after exiting.
func (i *Instance) Run(ctx context.Context) error {
	level.Debug(i.logger).Log("msg", "running instance", "name", i.cfg.Name)

	// trackingReg wraps the register for the instance to make sure that if Run
	// exits, any metrics Prometheus registers are removed and can be
	// re-registered if Run is called again.
	trackingReg := unregisterAllRegisterer{wrap: i.reg}
	defer trackingReg.UnregisterAll()

	wstore, err := i.newWal(&trackingReg)
	if err != nil {
		return err
	}

	discovery, err := i.newDiscoveryManager(ctx)
	if err != nil {
		return err
	}

	storage, err := i.newStorage(&trackingReg, wstore, i.readyScrapeManager)
	if err != nil {
		return err
	}

	scrapeManager := newScrapeManager(log.With(i.logger, "component", "scrape manager"), storage)

	err = scrapeManager.ApplyConfig(&config.Config{
		GlobalConfig:  i.globalCfg,
		ScrapeConfigs: i.cfg.ScrapeConfigs,
	})
	if err != nil {
		level.Error(i.logger).Log("msg", "failed applying config to scrape manager", "err", err)
		return fmt.Errorf("failed applying config to scrape manager: %w", err)
	}
	i.readyScrapeManager.Set(scrapeManager)

	rg := runGroupWithContext(ctx)

	// The actors defined here are defined in the order we want them to shut down.
	// Primarily, we want to ensure that the following shutdown order is
	// maintained:
	//		1. The scrape manager stops
	//    2. WAL storage is closed
	//    3. Remote write storage is closed
	// This is done to allow the instance to write stale markers for all active
	// series.
	{
		// Target Discovery
		rg.Add(discovery.Run, discovery.Stop)
	}
	{
		// Truncation loop
		ctx, contextCancel := context.WithCancel(context.Background())
		defer contextCancel()
		rg.Add(
			func() error {
				i.truncateLoop(ctx, wstore)
				level.Info(i.logger).Log("msg", "truncation loop stopped")
				return nil
			},
			func(err error) {
				level.Info(i.logger).Log("msg", "stopping truncation loop...")
				contextCancel()
			},
		)
	}
	{
		// Scrape manager
		rg.Add(
			func() error {
				err := scrapeManager.Run(discovery.SyncCh())
				level.Info(i.logger).Log("msg", "scrape manager stopped")
				return err
			},
			func(err error) {
				// The scrape manager is closed first to allow us to write staleness
				// markers without receiving new samples from scraping in the meantime.
				level.Info(i.logger).Log("msg", "stopping scrape manager...")
				scrapeManager.Stop()

				// On a graceful shutdown, write staleness markers. If something went
				// wrong, then the instance will be relaunched.
				if err == nil && i.cfg.WriteStaleOnShutdown {
					level.Info(i.logger).Log("msg", "writing staleness markers...")
					err := wstore.WriteStalenessMarkers(i.getRemoteWriteTimestamp)
					if err != nil {
						level.Error(i.logger).Log("msg", "error writing staleness markers", "err", err)
					}
				}

				level.Info(i.logger).Log("msg", "closing storage...")
				if err := storage.Close(); err != nil {
					level.Error(i.logger).Log("msg", "error stopping storage", "err", err)
				}
			},
		)
	}

	err = rg.Run()
	if err != nil {
		level.Error(i.logger).Log("msg", "agent instance stopped with error", "err", err)
	}
	return err
}

// TargetsActive returns the set of active targets from the scrape manager. Returns nil
// if the scrape manager is not ready yet.
func (i *Instance) TargetsActive() map[string][]*scrape.Target {
	mgr, err := i.readyScrapeManager.Get()
	if err == scrape.ErrNotReady {
		return nil
	} else if err != nil {
		level.Error(i.logger).Log("msg", "failed to get scrape manager when collecting active targets", "err", err)
		return nil
	}
	return mgr.TargetsActive()
}

type discoveryService struct {
	RunFunc    func() error
	StopFunc   func(err error)
	SyncChFunc func() GroupChannel
}

func (s *discoveryService) Run() error           { return s.RunFunc() }
func (s *discoveryService) Stop(err error)       { s.StopFunc(err) }
func (s *discoveryService) SyncCh() GroupChannel { return s.SyncChFunc() }

// newDiscoveryManager returns an implementation of a runnable service
// that outputs discovered targets to a channel. The implementation
// uses the Prometheus Discovery Manager. Targets will be filtered
// if the instance is configured to perform host filtering.
func (i *Instance) newDiscoveryManager(ctx context.Context) (*discoveryService, error) {
	ctx, cancel := context.WithCancel(ctx)

	logger := log.With(i.logger, "component", "discovery manager")
	manager := discovery.NewManager(ctx, logger, discovery.Name("scrape"))

	// TODO(rfratto): refactor this to a function?
	// TODO(rfratto): ensure job name name is unique
	c := map[string]sd_config.ServiceDiscoveryConfig{}
	for _, v := range i.cfg.ScrapeConfigs {
		c[v.JobName] = v.ServiceDiscoveryConfig
	}
	err := manager.ApplyConfig(c)
	if err != nil {
		cancel()
		level.Error(i.logger).Log("msg", "failed applying config to discovery manager", "err", err)
		return nil, fmt.Errorf("failed applying config to discovery manager: %w", err)
	}

	rg := runGroupWithContext(ctx)

	// Run the manager
	rg.Add(func() error {
		err := manager.Run()
		level.Info(i.logger).Log("msg", "discovery manager stopped")
		return err
	}, func(err error) {
		level.Info(i.logger).Log("msg", "stopping discovery manager...")
		cancel()
	})

	syncChFunc := manager.SyncCh

	// If host filtering is enabled, run it and use its channel for discovered
	// targets.
	if i.cfg.HostFilter {
		filterer, err := i.newHostFilter()
		if err != nil {
			cancel()
			return nil, err
		}

		rg.Add(func() error {
			filterer.Run(manager.SyncCh())
			level.Info(i.logger).Log("msg", "host filterer stopped")
			return nil
		}, func(_ error) {
			level.Info(i.logger).Log("msg", "stopping host filterer...")
			filterer.Stop()
		})

		syncChFunc = filterer.SyncCh
	}

	return &discoveryService{
		RunFunc:    rg.Run,
		StopFunc:   rg.Stop,
		SyncChFunc: syncChFunc,
	}, nil
}

func (i *Instance) newStorage(reg prometheus.Registerer, wal walStorage, sm scrape.ReadyManager) (storage.Storage, error) {
	logger := log.With(i.logger, "component", "remote")

	store := remote.NewStorage(logger, reg, wal.StartTime, wal.Directory(), i.cfg.RemoteFlushDeadline, sm)
	err := store.ApplyConfig(&config.Config{
		GlobalConfig:       i.globalCfg,
		RemoteWriteConfigs: i.cfg.RemoteWrite,
	})
	if err != nil {
		level.Error(i.logger).Log("msg", "failed applying config to remote storage", "err", err)
		return nil, fmt.Errorf("failed applying config to remote storage: %w", err)
	}

	fanoutStorage := storage.NewFanout(i.logger, wal, store)
	return fanoutStorage, nil
}

func (i *Instance) newHostFilter() (*HostFilter, error) {
	hostname, err := Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to create host filterer: %w", err)
	}

	level.Debug(i.logger).Log("msg", "creating host filterer", "for_host", hostname, "enabled", i.cfg.HostFilter)
	return NewHostFilter(hostname), nil
}

func (i *Instance) truncateLoop(ctx context.Context, wal walStorage) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(i.cfg.WALTruncateFrequency):
			ts := i.getRemoteWriteTimestamp()
			if ts == 0 {
				level.Debug(i.logger).Log("msg", "can't truncate the WAL yet")
				continue
			}

			level.Debug(i.logger).Log("msg", "truncating the WAL", "ts", ts)
			err := wal.Truncate(ts)
			if err != nil {
				// The only issue here is larger disk usage and a greater replay time,
				// so we'll only log this as a warning.
				level.Warn(i.logger).Log("msg", "could not truncate WAL", "err", err)
			}
		}
	}
}

// getRemoteWriteTimestamp looks up the last successful remote write timestamp.
// This is passed to wal.Storage for its truncation. If no remote write sections
// are configured, getRemoteWriteTimestamp returns the current time.
func (i *Instance) getRemoteWriteTimestamp() int64 {
	if len(i.cfg.RemoteWrite) == 0 {
		return timestamp.FromTime(time.Now())
	}

	lbls := make([]string, len(i.cfg.RemoteWrite))
	for idx := 0; idx < len(lbls); idx++ {
		lbls[idx] = i.cfg.RemoteWrite[idx].Name
	}

	vals, err := i.vc.GetValues("remote_name", lbls...)
	if err != nil {
		level.Error(i.logger).Log("msg", "could not get remote write timestamps", "err", err)
		return 0
	}
	if len(vals) == 0 {
		return 0
	}

	// We use the lowest value from the metric since we don't want to delete any
	// segments from the WAL until they've been written by all of the remote_write
	// configurations.
	ts := int64(math.MaxInt64)
	for _, val := range vals {
		ival := int64(val)
		if ival < ts {
			ts = ival
		}
	}

	// Convert to the millisecond precision which is used by the WAL
	return ts * 1000
}

// walStorage is an interface satisfied by wal.Storage, and created for testing.
type walStorage interface {
	// walStorage implements Queryable for compatibility, but is unused.
	storage.Queryable

	Directory() string

	StartTime() (int64, error)
	WriteStalenessMarkers(remoteTsFunc func() int64) error
	Appender() storage.Appender
	Truncate(mint int64) error

	Close() error
}

type unregisterAllRegisterer struct {
	wrap prometheus.Registerer
	cs   map[prometheus.Collector]struct{}
}

// Register implements prometheus.Registerer.
func (u *unregisterAllRegisterer) Register(c prometheus.Collector) error {
	if u.wrap == nil {
		return nil
	}

	err := u.wrap.Register(c)
	if err != nil {
		return err
	}
	if u.cs == nil {
		u.cs = make(map[prometheus.Collector]struct{})
	}
	u.cs[c] = struct{}{}
	return nil
}

// MustRegister implements prometheus.Registerer.
func (u *unregisterAllRegisterer) MustRegister(cs ...prometheus.Collector) {
	for _, c := range cs {
		if err := u.Register(c); err != nil {
			panic(err)
		}
	}
}

// Unregister implements prometheus.Registerer.
func (u *unregisterAllRegisterer) Unregister(c prometheus.Collector) bool {
	if u.wrap == nil {
		return false
	}
	ok := u.wrap.Unregister(c)
	if ok && u.cs != nil {
		delete(u.cs, c)
	}
	return ok
}

// UnregisterAll unregisters all collectors that were registered through the
// Reigsterer.
func (u *unregisterAllRegisterer) UnregisterAll() {
	if u.cs == nil {
		return
	}
	for c := range u.cs {
		u.Unregister(c)
	}
}

// Hostname retrieves the hostname identifying the machine the process is
// running on. It will return the value of $HOSTNAME, if defined, and fall
// back to Go's os.Hostname.
func Hostname() (string, error) {
	hostname := os.Getenv("HOSTNAME")
	if hostname != "" {
		return hostname, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return hostname, nil
}

func getHash(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:]), nil
}

// MetricValueCollector wraps around a Gatherer and provides utilities for
// pulling metric values from a given metric name and label matchers.
//
// This is used by the agent instances to find the most recent timestamp
// successfully remote_written to for pruposes of safely truncating the WAL.
//
// MetricValueCollector is only intended for use with Gauges and Counters.
type MetricValueCollector struct {
	g     prometheus.Gatherer
	match string
}

// NewMetricValueCollector creates a new MetricValueCollector.
func NewMetricValueCollector(g prometheus.Gatherer, match string) *MetricValueCollector {
	return &MetricValueCollector{
		g:     g,
		match: match,
	}
}

// GetValues looks through all the tracked metrics and returns all values
// for metrics that match some key value pair.
func (vc *MetricValueCollector) GetValues(label string, labelValues ...string) ([]float64, error) {
	vals := []float64{}

	families, err := vc.g.Gather()
	if err != nil {
		return nil, err
	}

	for _, family := range families {
		if !strings.Contains(family.GetName(), vc.match) {
			continue
		}

		for _, m := range family.GetMetric() {
			matches := false
			for _, l := range m.GetLabel() {
				if l.GetName() != label {
					continue
				}

				v := l.GetValue()
				for _, match := range labelValues {
					if match == v {
						matches = true
						break
					}
				}
				break
			}
			if !matches {
				continue
			}

			var value float64
			if m.Gauge != nil {
				value = m.Gauge.GetValue()
			} else if m.Counter != nil {
				value = m.Counter.GetValue()
			} else if m.Untyped != nil {
				value = m.Untyped.GetValue()
			} else {
				return nil, errors.New("tracking unexpected metric type")
			}

			vals = append(vals, value)
		}
	}

	return vals, nil
}

func newScrapeManager(logger log.Logger, app storage.Appendable) *scrape.Manager {
	// scrape.NewManager modifies a global variable in Prometheus. To avoid a
	// data race of modifying that global, we lock a mutex here briefly.
	managerMtx.Lock()
	defer managerMtx.Unlock()
	return scrape.NewManager(logger, app)
}

type runGroupContext struct {
	cancel context.CancelFunc

	g *run.Group
}

// runGroupWithContext creates a new run.Group that will be stopped if the
// context gets canceled in addition to the normal behavior of stopping
// when any of the actors stop.
func runGroupWithContext(ctx context.Context) *runGroupContext {
	ctx, cancel := context.WithCancel(ctx)

	var g run.Group
	g.Add(func() error {
		<-ctx.Done()
		return nil
	}, func(_ error) {
		cancel()
	})

	return &runGroupContext{cancel: cancel, g: &g}
}

func (rg *runGroupContext) Add(execute func() error, interrupt func(error)) {
	rg.g.Add(execute, interrupt)
}

func (rg *runGroupContext) Run() error   { return rg.g.Run() }
func (rg *runGroupContext) Stop(_ error) { rg.cancel() }
