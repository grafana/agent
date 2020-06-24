// Package loki implements Loki logs support for the Grafana Cloud Agent.
package loki

import (
	"flag"

	"github.com/grafana/loki/pkg/promtail"
	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/grafana/loki/pkg/promtail/config"
	"github.com/grafana/loki/pkg/promtail/positions"
	"github.com/grafana/loki/pkg/promtail/scrape"
	"github.com/grafana/loki/pkg/promtail/server"
	"github.com/grafana/loki/pkg/promtail/targets"
)

// Config controls the configuration of the Loki log scraper.
type Config struct {
	ClientConfigs   []client.Config  `yaml:"clients,omitempty"`
	PositionsConfig positions.Config `yaml:"positions,omitempty"`
	ScrapeConfig    []scrape.Config  `yaml:"scrape_configs,omitempty"`
	TargetConfig    targets.Config   `yaml:"target_config,omitempty"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	// TODO(rfratto): make this a RegisterFlagsWithPrefix function
	c.PositionsConfig.RegisterFlags(f)
	c.TargetConfig.RegisterFlags(f)
}

type Loki struct {
	p *promtail.Promtail
}

// New creates and starts Loki log collection.
func New(c Config) (*Loki, error) {
	p, err := promtail.New(config.Config{
		ServerConfig:    server.Config{Disable: true},
		ClientConfigs:   c.ClientConfigs,
		PositionsConfig: c.PositionsConfig,
		ScrapeConfig:    c.ScrapeConfig,
		TargetConfig:    c.TargetConfig,
	}, false)
	if err != nil {
		return nil, err
	}

	return &Loki{p: p}, nil
}

func (l *Loki) Stop() {
	l.p.Shutdown()
}
