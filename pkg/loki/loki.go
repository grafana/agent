// Package loki implements Loki logs support for the Grafana Cloud Agent.
package loki

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/loki/pkg/promtail"
	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/grafana/loki/pkg/promtail/config"
	"github.com/grafana/loki/pkg/promtail/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

func init() {
	client.UserAgent = fmt.Sprintf("GrafanaCloudAgent/%s", version.Version)
}

type Loki struct {
	promtails []*promtail.Promtail
}

// New creates and starts Loki log collection.
func New(reg prometheus.Registerer, c Config, l log.Logger) (*Loki, error) {
	l = log.With(l, "component", "loki")

	if c.PositionsDirectory != "" {
		err := os.MkdirAll(c.PositionsDirectory, 0700)
		if err != nil {
			level.Warn(l).Log("msg", "failed to create the positions directory. logs may be unable to save their position", "path", c.PositionsDirectory, "err", err)
		}
	}

	promtails := make([]*promtail.Promtail, 0, len(c.Configs))

	for _, ic := range c.Configs {
		if len(ic.ClientConfigs) == 0 {
			level.Info(l).Log("msg", "skipping creation of a promtail because no client_configs are present", "config", ic.Name)
			continue
		}

		r := prometheus.WrapRegistererWith(prometheus.Labels{"loki_name": ic.Name}, reg)

		p, err := promtail.New(config.Config{
			ServerConfig:    server.Config{Disable: true},
			ClientConfigs:   ic.ClientConfigs,
			PositionsConfig: ic.PositionsConfig,
			ScrapeConfig:    ic.ScrapeConfig,
			TargetConfig:    ic.TargetConfig,
		}, false, promtail.WithLogger(l), promtail.WithRegisterer(r))
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
