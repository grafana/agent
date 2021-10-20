package frontend_receiver

import (
	"context"
	"log"
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/exporters"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/models"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/receiver"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

type Integration struct {
	receiver *receiver.FrontendReceiver
	logger   kitlog.Logger
	config   *Config
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

func (i *Integration) Handlers() (map[string]http.Handler, error) {
	logger := log.New(kitlog.NewStdlibAdapter(i.logger), "frontend_receiver:collector", 0)
	return map[string]http.Handler{
		i.config.Endpoint: i.receiver.ReceiverHandler(*logger),
	}, nil
}

func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{}
}

func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}

func New(log kitlog.Logger, c *Config) (integrations.Integration, error) {
	level.Debug(log).Log("msg", "initializing frontend receiver", "config", c)

	exporters := []exporters.FrontendReceiverExporter{
		func(payload models.Payload) error {

			return nil
		},
	}

	receiver := receiver.NewFrontendReceiver(c.Receiver, exporters)

	integration := &Integration{
		receiver: &receiver,
		config:   c,
	}

	return integration, nil
}
