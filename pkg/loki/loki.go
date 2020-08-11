// Package loki implements Loki logs support for the Grafana Cloud Agent.
package loki

import (
	"flag"

	"github.com/go-kit/kit/log"
	"github.com/grafana/loki/pkg/promtail"
	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/grafana/loki/pkg/promtail/config"
	"github.com/grafana/loki/pkg/promtail/positions"
	"github.com/grafana/loki/pkg/promtail/scrapeconfig"
	"github.com/grafana/loki/pkg/promtail/server"
	"github.com/grafana/loki/pkg/promtail/targets/file"
)

// Config controls the configuration of the Loki log scraper.
type Config struct {
	ClientConfigs   []client.Config       `yaml:"clients,omitempty"`
	PositionsConfig positions.Config      `yaml:"positions,omitempty"`
	ScrapeConfig    []scrapeconfig.Config `yaml:"scrape_configs,omitempty"`
	TargetConfig    file.Config           `yaml:"target_config,omitempty"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.PositionsConfig.RegisterFlagsWithPrefix("loki.", f)
	c.TargetConfig.RegisterFlagsWithPrefix("loki.", f)
}

type Loki struct {
	p *promtail.Promtail
}

// New creates and starts Loki log collection.
func New(c Config, l log.Logger) (*Loki, error) {
	l = log.With(l, "component", "loki")

	p, err := promtail.New(config.Config{
		ServerConfig:    server.Config{Disable: true},
		ClientConfigs:   c.ClientConfigs,
		PositionsConfig: c.PositionsConfig,
		ScrapeConfig:    c.ScrapeConfig,
		TargetConfig:    c.TargetConfig,
	}, false, promtail.WithLogger(l))
	if err != nil {
		return nil, err
	}

	return &Loki{p: p}, nil
}

func (l *Loki) Stop() {
	l.p.Shutdown()
}
