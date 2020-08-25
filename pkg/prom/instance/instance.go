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
	"github.com/grafana/loki/pkg/build"
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

func init() {
	remote.UserAgent = fmt.Sprintf("GrafanaCloudAgent/%s", build.Version)
}

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
	// All fields in the following block may be accessed and modified by
	// concurrently running goroutines.
	//
	// Note that all Prometheus components listed here may be nil at any
	// given time; methods reading them should take care to do nil checks.
	mut                sync.Mutex
	cfg                Config
	wal                walStorage
	discovery          *discoveryService
	readyScrapeManager *scrape.ReadyScrapeManager
	remoteStore        *remote.Storage
	storage            storage.Storage

	globalCfg config.GlobalConfig
	logger    log.Logger

	reg    prometheus.Registerer
	newWal walStorageFactory

	vc *MetricValueCollector
}

// New creates a new Instance with a directory for storing the WAL. The instance
// will not start until Run is called on the instance.
func New(reg prometheus.Registerer, globalCfg config.GlobalConfig, cfg Config, walDir string, logger log.Logger) (*Instance, error) {
	logger = log.With(logger, "instance", cfg.Name)

	instWALDir := filepath.Join(walDir, cfg.Name)

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

// Run starts the instance, initializing Prometheus components, and will
// continue to run until an error happens during execution or the provided
// context is cancelled.
//
// Run may be re-called after exiting, as components will be reinitialized each
// time Run is called.
func (i *Instance) Run(ctx context.Context) error {
	// i.cfg may change at any point in the middle of this method but not in a way
	// that affects any of the code below; rather than grabbing a mutex every time
	// we want to read the config, we'll simplify the access and just grab a copy
	// now.
	i.mut.Lock()
	cfg := i.cfg
	i.mut.Unlock()

	level.Debug(i.logger).Log("msg", "initializing instance", "name", cfg.Name)

	// trackingReg wraps the register for the instance to make sure that if Run
	// exits, any metrics Prometheus registers are removed and can be
	// re-registered if Run is called again.
	trackingReg := unregisterAllRegisterer{wrap: i.reg}
	defer trackingReg.UnregisterAll()

	if err := i.initialize(ctx, &trackingReg, &cfg); err != nil {
		level.Error(i.logger).Log("msg", "failed to initialize instance", "err", err)
		return fmt.Errorf("failed to initialize instance: %w", err)
	}

	// The actors defined here are defined in the order we want them to shut down.
	// Primarily, we want to ensure that the following shutdown order is
	// maintained:
	//		1. The scrape manager stops
	//    2. WAL storage is closed
	//    3. Remote write storage is closed
	// This is done to allow the instance to write stale markers for all active
	// series.
	rg := runGroupWithContext(ctx)

	{
		// Target Discovery
		rg.Add(i.discovery.Run, i.discovery.Stop)
	}
	{
		// Truncation loop
		ctx, contextCancel := context.WithCancel(context.Background())
		defer contextCancel()
		rg.Add(
			func() error {
				i.truncateLoop(ctx, i.wal, &cfg)
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
		sm, err := i.readyScrapeManager.Get()
		if err != nil {
			level.Error(i.logger).Log("msg", "failed to get scrape manager")
			return err
		}

		// Scrape manager
		rg.Add(
			func() error {
				err := sm.Run(i.discovery.SyncCh())
				level.Info(i.logger).Log("msg", "scrape manager stopped")
				return err
			},
			func(err error) {
				// The scrape manager is closed first to allow us to write staleness
				// markers without receiving new samples from scraping in the meantime.
				level.Info(i.logger).Log("msg", "stopping scrape manager...")
				sm.Stop()

				// On a graceful shutdown, write staleness markers. If something went
				// wrong, then the instance will be relaunched.
				if err == nil && cfg.WriteStaleOnShutdown {
					level.Info(i.logger).Log("msg", "writing staleness markers...")
					err := i.wal.WriteStalenessMarkers(i.getRemoteWriteTimestamp)
					if err != nil {
						level.Error(i.logger).Log("msg", "error writing staleness markers", "err", err)
					}
				}

				level.Info(i.logger).Log("msg", "closing storage...")
				if err := i.storage.Close(); err != nil {
					level.Error(i.logger).Log("msg", "error stopping storage", "err", err)
				}
			},
		)
	}

	level.Debug(i.logger).Log("msg", "running instance", "name", cfg.Name)
	err := rg.Run()
	if err != nil {
		level.Error(i.logger).Log("msg", "agent instance stopped with error", "err", err)
	}
	return err
}

// initialize sets up the various Prometheus components with their initial
// settings. initialize will be called each time the Instance is run. Prometheus
// components cannot be reused after they are stopped so we need to recreate them
// each run.
func (i *Instance) initialize(ctx context.Context, reg prometheus.Registerer, cfg *Config) error {
	i.mut.Lock()
	defer i.mut.Unlock()

	var err error

	i.wal, err = i.newWal(reg)
	if err != nil {
		return fmt.Errorf("error creating WAL: %w", err)
	}

	i.discovery, err = i.newDiscoveryManager(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error creating discovery manager: %w", err)
	}

	i.readyScrapeManager = &scrape.ReadyScrapeManager{}

	// Setup the remote storage
	remoteLogger := log.With(i.logger, "component", "remote")
	i.remoteStore = remote.NewStorage(remoteLogger, reg, i.wal.StartTime, i.wal.Directory(), cfg.RemoteFlushDeadline, i.readyScrapeManager)
	err = i.remoteStore.ApplyConfig(&config.Config{
		GlobalConfig:       i.globalCfg,
		RemoteWriteConfigs: cfg.RemoteWrite,
	})
	if err != nil {
		return fmt.Errorf("failed applying config to remote storage: %w", err)
	}

	i.storage = storage.NewFanout(i.logger, i.wal, i.remoteStore)

	scrapeManager := newScrapeManager(log.With(i.logger, "component", "scrape manager"), i.storage)
	err = scrapeManager.ApplyConfig(&config.Config{
		GlobalConfig:  i.globalCfg,
		ScrapeConfigs: cfg.ScrapeConfigs,
	})
	if err != nil {
		return fmt.Errorf("failed applying config to scrape manager: %w", err)
	}

	i.readyScrapeManager.Set(scrapeManager)

	return nil
}

// Update accepts a new Config for the Instance and will dynamically update any
// running Prometheus components with the new values from Config. Update will
// return an ErrInvalidUpdate if the Update could not be applied.
func (i *Instance) Update(c Config) error {
	i.mut.Lock()
	defer i.mut.Unlock()

	// It's only (currently) valid to update scrape_configs and remote_write, so
	// if any other field has changed here, return the error.
	var err error
	switch {
	// This first case should never happen in practice but it's included here for
	// completions sake.
	case i.cfg.Name != c.Name:
		err = errImmutableField{Field: "name"}
	case i.cfg.HostFilter != c.HostFilter:
		err = errImmutableField{Field: "host_filter"}
	case i.cfg.WALTruncateFrequency != c.WALTruncateFrequency:
		err = errImmutableField{Field: "wal_truncate_frequency"}
	case i.cfg.RemoteFlushDeadline != c.RemoteFlushDeadline:
		err = errImmutableField{Field: "remote_flush_deadline"}
	case i.cfg.WriteStaleOnShutdown != c.WriteStaleOnShutdown:
		err = errImmutableField{Field: "write_stale_on_shutdown"}
	}
	if err != nil {
		return ErrInvalidUpdate{Inner: err}
	}

	// Check to see if the components exist yet.
	if i.discovery == nil || i.remoteStore == nil || i.readyScrapeManager == nil {
		return ErrInvalidUpdate{
			Inner: fmt.Errorf("cannot dynamically update because instance is not running"),
		}
	}

	sdConfigs := map[string]sd_config.ServiceDiscoveryConfig{}
	for _, v := range c.ScrapeConfigs {
		sdConfigs[v.JobName] = v.ServiceDiscoveryConfig
	}
	err = i.discovery.Manager.ApplyConfig(sdConfigs)
	if err != nil {
		return fmt.Errorf("failed applying configs to discovery manager: %w", err)
	}

	err = i.remoteStore.ApplyConfig(&config.Config{
		GlobalConfig:       i.globalCfg,
		RemoteWriteConfigs: c.RemoteWrite,
	})
	if err != nil {
		return fmt.Errorf("error applying new remote_write configs: %w", err)
	}

	sm, err := i.readyScrapeManager.Get()
	if err != nil {
		return fmt.Errorf("couldn't get scrape manager to apply new scrape configs: %w", err)
	}

	err = sm.ApplyConfig(&config.Config{
		GlobalConfig:  i.globalCfg,
		ScrapeConfigs: c.ScrapeConfigs,
	})
	if err != nil {
		return fmt.Errorf("error applying updated configs to scrape manager: %w", err)
	}

	i.cfg = c
	return nil
}

// TargetsActive returns the set of active targets from the scrape manager. Returns nil
// if the scrape manager is not ready yet.
func (i *Instance) TargetsActive() map[string][]*scrape.Target {
	i.mut.Lock()
	defer i.mut.Unlock()

	if i.readyScrapeManager == nil {
		return nil
	}

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
	Manager *discovery.Manager

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
func (i *Instance) newDiscoveryManager(ctx context.Context, cfg *Config) (*discoveryService, error) {
	ctx, cancel := context.WithCancel(ctx)

	logger := log.With(i.logger, "component", "discovery manager")
	manager := discovery.NewManager(ctx, logger, discovery.Name("scrape"))

	// TODO(rfratto): refactor this to a function?
	// TODO(rfratto): ensure job name name is unique
	c := map[string]sd_config.ServiceDiscoveryConfig{}
	for _, v := range cfg.ScrapeConfigs {
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
	if cfg.HostFilter {
		hostname, err := Hostname()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create host filterer: %w", err)
		}
		level.Debug(i.logger).Log("msg", "creating host filterer", "for_host", hostname)

		filterer := NewHostFilter(hostname)

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
		Manager: manager,

		RunFunc:    rg.Run,
		StopFunc:   rg.Stop,
		SyncChFunc: syncChFunc,
	}, nil
}

func (i *Instance) truncateLoop(ctx context.Context, wal walStorage, cfg *Config) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(cfg.WALTruncateFrequency):
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
	i.mut.Lock()
	defer i.mut.Unlock()

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
	// walStorage implements Queryable/ChunkQueryable for compatibility, but is unused.
	storage.Queryable
	storage.ChunkQueryable

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
