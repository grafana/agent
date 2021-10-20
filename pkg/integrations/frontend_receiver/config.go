package frontend_receiver

import (
	"github.com/grafana/agent/pkg/integrations/config"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/tempo"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	recconf "github.com/grafana/grafana-frontend-telemetry-receiver/pkg/config"
)

type Config struct {
	Common   config.Common          `yaml:",inline"`
	Receiver recconf.ReceiverConfig `yaml:",inline"`
	Endpoint string                 `yaml:"endpoint"`
}

func (c *Config) CommonConfig() config.Common {
	return c.Common
}

func (c *Config) Name() string {
	return "frontend_receiver"
}

func (c *Config) NewIntegration(l log.Logger, loki *loki.Logs, tempo *tempo.Tempo) (integrations.Integration, error) {
	return New(l, c)
}
