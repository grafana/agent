// Package loki implements Loki logs support for the Grafana Cloud Agent.
package loki

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/loki/pkg/promtail"
	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/grafana/loki/pkg/promtail/config"
	"github.com/grafana/loki/pkg/promtail/server"
	"github.com/prometheus/common/version"
)

func init() {
	client.UserAgent = fmt.Sprintf("GrafanaCloudAgent/%s", version.Version)
}

type Loki struct {
	promtails []*promtail.Promtail
}

// New creates and starts Loki log collection.
func New(c Config, l log.Logger) (*Loki, error) {
	l = log.With(l, "component", "loki")

	if c.Version == "" {
		level.Warn(l).Log("msg", "no Loki version field detected, defaulting to v0. the default will change in a future release!")
	}

	promtails := make([]*promtail.Promtail, 0, len(c.Config.Configs))

	for _, ic := range c.Config.Configs {
		if len(ic.ClientConfigs) == 0 {
			level.Info(l).Log("msg", "skipping creation of a promtail because no client_configs are present", "config", ic.Name)
			continue
		}

		p, err := promtail.New(config.Config{
			ServerConfig:    server.Config{Disable: true},
			ClientConfigs:   ic.ClientConfigs,
			PositionsConfig: ic.PositionsConfig,
			ScrapeConfig:    ic.ScrapeConfig,
			TargetConfig:    ic.TargetConfig,
		}, false, promtail.WithLogger(l))
		if err != nil {
			return nil, err
		}

		promtails = append(promtails, p)
	}

	return &Loki{promtails: promtails}, nil
}

func (l *Loki) Stop() {
	for _, p := range l.promtails {
		p.Shutdown()
	}
}
