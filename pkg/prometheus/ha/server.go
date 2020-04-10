// Package ha implements a high availability clustering mode for the agent. It
// is also referred to as the "scraping service" mode, as this is a fairly
// accurate description of what it does: a series of configs are stored in a
// KV store and a cluster of agents pulls configs from the store and shards
// them amongst the cluster, thereby distributing scraping load.
package ha

import (
	"context"
	"flag"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prometheus/instance"
)

// ConfigManager is an interface to manipulating a set of running
// instance.Configs. It is satisfied by the ConfigManager struct in
// pkg/prometheus, but is provided as an interface here for testing and
// avoiding import cycles.
type ConfigManager interface {
	// ListConfigs gets the list of currently known configs.
	ListConfigs() map[string]instance.Config

	// ApplyConfig adds or updates a config.
	ApplyConfig(c instance.Config)

	// DeleteConfig deletes a config by name, uniquely keyed by the
	// Name field in instance.Config.
	DeleteConfig(name string) error
}

// Config describes how to instantiate a scraping service Server instance.
type Config struct {
	Enabled    bool                  `yaml:"enabled"`
	KVStore    kv.Config             `yaml:"kvstore"`
	Lifecycler ring.LifecyclerConfig `yaml:"lifecycler"`
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
	c.KVStore.RegisterFlagsWithPrefix(prefix+"config-store.", "configurations/", f)
	c.Lifecycler.RegisterFlagsWithPrefix(prefix, f)
}

// Server implements the HA scraping service.
type Server struct {
	cfg    Config
	logger log.Logger

	cm ConfigManager

	// Interface to storing and retreiving config objects
	kv kv.Client

	// Management for being a cluster member
	lc *ring.Lifecycler

	// View into members of the cluster
	ring *ring.Ring
}

// New creates a new HA scraping service instance.
func New(cfg Config, logger log.Logger, cm ConfigManager) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		logger: log.With(logger, "component", "ha"),
	}

	var err error
	s.kv, err = kv.NewClient(cfg.KVStore, GetCodec())
	if err != nil {
		return nil, err
	}

	s.lc, err = ring.NewLifecycler(cfg.Lifecycler, nil, "agent", "agent", true)
	if err != nil {
		return nil, err
	}
	s.lc.StartAsync(context.Background())

	s.ring, err = ring.New(cfg.Lifecycler.RingConfig, "agent_viewer", "agent")
	if err != nil {
		return nil, err
	}
	s.ring.StartAsync(context.Background())

	return s, nil
}

// Stop stops the HA server and its dependencies.
func (s *Server) Stop() error {
	s.ring.StopAsync()
	ringErr := s.ring.AwaitTerminated(context.Background())

	s.lc.StopAsync()
	lcErr := s.lc.AwaitTerminated(context.Background())

	// TODO(rfratto): combine errors?
	if ringErr != nil {
		return ringErr
	}

	return lcErr
}
