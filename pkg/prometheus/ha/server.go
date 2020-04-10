// Package ha implements a high availability clustering mode for the agent. It
// is also referred to as the "scraping service" mode, as this is a fairly
// accurate description of what it does: a series of configs are stored in a
// KV store and a cluster of agents pulls configs from the store and shards
// them amongst the cluster, thereby distributing scraping load.
package ha

import (
	"flag"

	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
)

// Config describes how to instantiate a scraping service Server instance.
type Config struct {
	Enabled bool      `yaml:"enabled"`
	KVStore kv.Config `yaml:"kvstore"`
}

// RegisterFlagsWithPrefix adds the flags required to config this to the given
// FlagSet with a specified prefix.
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"enabled", false, "enables the scraping service mode")
	c.KVStore.RegisterFlagsWithPrefix(prefix, "configurations/", f)
}

// Server implements the HA scraping service.
type Server struct {
	cfg    Config
	logger log.Logger

	kv kv.Client
}

// New creates a new HA scraping service instance.
func New(cfg Config, logger log.Logger) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		logger: log.With(logger, "component", "ha"),
	}

	var err error
	s.kv, err = kv.NewClient(cfg.KVStore, GetCodec())
	if err != nil {
		return nil, err
	}

	return s, nil
}
