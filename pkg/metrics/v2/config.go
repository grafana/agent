package metrics

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/cluster"
	prom_config "github.com/prometheus/prometheus/config"
)

// Defaults.
var (
	DefaultOptions = Options{
		WALDir:               "agent-data",
		WALTruncateFrequency: 2 * time.Hour,
		WALMinTime:           5 * time.Minute,
		WALMaxTime:           8 * time.Hour,

		RemoteFlushDeadline: 1 * time.Minute,
	}

	DefaultConfig = Config{
		Global: DefaultGlobalInstanceConfig,
	}

	DefaultGlobalInstanceConfig = GlobalInstanceConfig{
		Prometheus: prom_config.DefaultGlobalConfig,
	}

	DefaultInstanceConfig = InstanceConfig{}
)

// Options contain static options for instantiating the metrics subsystem.
// Options may not change between instantiations of the subsystem.
type Options struct {
	// Directory to store WAL data. If this is empty, the metrics subsystem
	// cannot be used and will fail to load any non-empty list of configs.
	WALDir string

	WALTruncateFrequency time.Duration // Frequency to truncate WAL data.
	WALMinTime           time.Duration // Minimum time data must exist in WAL before being garbage collected.
	WALMaxTime           time.Duration // Maximum time data may exist in WAL before being garbage collected.

	RemoteFlushDeadline time.Duration // How long to wait before giving up on flushing data on shutdown.

	// Cluster to use for distributing work. Must not be nil.
	Cluster *cluster.Node
}

func (o *Options) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.WALDir, "metrics.wal.directory", DefaultOptions.WALDir, "location to store metrics in a WAL")
	fs.DurationVar(&o.WALTruncateFrequency, "metrics.wal.truncate-frequency", DefaultOptions.WALTruncateFrequency, "frequency to perform GC on WALs")
	fs.DurationVar(&o.WALMinTime, "metrics.wal.min-live", DefaultOptions.WALMinTime, "time data must be alive before being deleteable")
	fs.DurationVar(&o.WALMaxTime, "metrics.wal.max-live", DefaultOptions.WALMaxTime, "time data can be alive before foricibly being deleted")
	fs.DurationVar(&o.RemoteFlushDeadline, "metrics.flush-deadline", DefaultOptions.RemoteFlushDeadline, "maximum time to wait to flush data on shutdown")
}

// Validate will ensure o is valid.
func (o *Options) Validate() error {
	switch {
	case o.WALDir == "":
		return fmt.Errorf("wal directory must be provided")
	case o.WALTruncateFrequency <= 0:
		return fmt.Errorf("wal_truncate_frequency must be greater than 0s")
	case o.RemoteFlushDeadline <= 0:
		return fmt.Errorf("remote_flush_deadline must be greater than 0s")
	case o.WALMinTime >= o.WALMaxTime:
		return fmt.Errorf("min_wal_time must be less than max_wal_time")
	default:
		return nil
	}
}

// Config holds runtime configuration for the metrics subsystem. New instances
// of Config can be passed at runtime.
type Config struct {
	Global  GlobalInstanceConfig `yaml:"global,omitempty"`
	Configs []InstanceConfig     `yaml:"configs,omitempty"`
}

// ApplyDefaults will apply runtime defaults to Config.
func (c *Config) ApplyDefaults(o Options) error {
	usedNames := map[string]struct{}{}

	for i := range c.Configs {
		name := c.Configs[i].Name
		if err := c.Configs[i].ApplyDefaults(c.Global, o); err != nil {
			// Try to show a helpful name in the error
			if name == "" {
				name = fmt.Sprintf("at index %d", i)
			}
			return fmt.Errorf("error validating instance %q: %w", name, err)
		}

		if _, ok := usedNames[name]; ok {
			return fmt.Errorf(
				"metrics instance names must be unique. found multiple instances with name %q",
				name,
			)
		}
		usedNames[name] = struct{}{}
	}

	return nil
}

// GlobalInstanceConfig is a set of configuration values shared amongst all
// metrics instances.
type GlobalInstanceConfig struct {
	Prometheus  prom_config.GlobalConfig         `yaml:",inline"`                // Defaults inherited from Prometheus
	RemoteWrite []*prom_config.RemoteWriteConfig `yaml:"remote_write,omitempty"` // Default remote_write settings
}

// InstanceConfig configures an individual metrics instance. Metrics instances
// combine collecting and sending metrics to a specific endpoint.
type InstanceConfig struct {
	// Name of the config. Must be unique across all metrics instances.
	Name          string                           `yaml:"name,omitempty"`
	ScrapeConfigs []*prom_config.ScrapeConfig      `yaml:"scrape_configs,omitempty"`
	RemoteWrite   []*prom_config.RemoteWriteConfig `yaml:"remote_write,omitempty"`
}

func (c *InstanceConfig) ApplyDefaults(g GlobalInstanceConfig, o Options) error {
	if c.Name == "" {
		return fmt.Errorf("missing metrics instance name")
	}

	jobNames := map[string]struct{}{}
	for _, sc := range c.ScrapeConfigs {
		if sc == nil {
			return fmt.Errorf("empty or null scrape config section")
		}

		// First set the correct scrape interval, then check that the timeout
		// (inferred or explicit) is not greater than that.
		if sc.ScrapeInterval == 0 {
			sc.ScrapeInterval = g.Prometheus.ScrapeInterval
		}
		if sc.ScrapeTimeout > sc.ScrapeInterval {
			return fmt.Errorf("scrape timeout greater than scrape interval for scrape config with job name %q", sc.JobName)
		}
		if time.Duration(sc.ScrapeInterval) > o.WALTruncateFrequency {
			return fmt.Errorf("scrape interval greater than wal_truncate_frequency for scrape config with job name %q", sc.JobName)
		}
		if sc.ScrapeTimeout == 0 {
			if g.Prometheus.ScrapeTimeout > sc.ScrapeInterval {
				sc.ScrapeTimeout = sc.ScrapeInterval
			} else {
				sc.ScrapeTimeout = g.Prometheus.ScrapeTimeout
			}
		}

		if _, exists := jobNames[sc.JobName]; exists {
			return fmt.Errorf("found multiple scrape configs with job name %q", sc.JobName)
		}
		jobNames[sc.JobName] = struct{}{}
	}

	rwNames := map[string]struct{}{}

	if len(c.RemoteWrite) == 0 {
		c.RemoteWrite = g.RemoteWrite
	}
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

func getHash(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:]), nil
}
