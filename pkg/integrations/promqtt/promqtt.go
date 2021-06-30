package promqtt

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/sh0rez/promqtt/relay"
)

func init() {
	integrations.RegisterIntegration(&Config{})
}

type Config struct {
	Common config.Common `yaml:",inline"`
	Relay  relay.Config  `yaml:",inline"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = Config{Relay: relay.DefaultConfig()}
	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "promqtt"
}

func (c *Config) CommonConfig() config.Common {
	return c.Common
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return &Integration{Config: c}, nil
}

type Integration struct {
	Config *Config
	Relay  *relay.Relay
}

func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.Config.Name(),
		MetricsPath: "/mqtt",
	}}
}

func (i *Integration) Run(ctx context.Context) error {
	var err error
	i.Relay, err = relay.New(i.Config.Relay)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

func (i *Integration) MetricsHandler() (http.Handler, error) {
	return i.Relay, nil
}
