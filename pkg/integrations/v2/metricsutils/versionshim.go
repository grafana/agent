package metricsutils

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/prometheus/common/model"

	v1 "github.com/grafana/agent/pkg/integrations"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
)

// CreateShim creates a shim between the v1.Config and v2.Config. The resulting
// config is NOT registered.
func CreateShim(before v1.Config) (after v2.UpgradedConfig) {
	return &configShim{Orig: before}
}

type configShim struct {
	Orig v1.Config
}

var (
	_ v2.Config         = (*configShim)(nil)
	_ v2.UpgradedConfig = (*configShim)(nil)
)

// LegacyConfig implements v2.UpgradedConfig.
func (s *configShim) LegacyConfig() v1.Config { return s.Orig }

func (s *configShim) Name() string { return s.Orig.Name() }
func (s *configShim) Identifier(g v2.Globals) (string, error) {
	return s.Orig.InstanceKey(g.AgentIdentifier)
}

func (s *configShim) NewIntegration(l log.Logger, g v2.Globals) (v2.Integration, error) {
	v1Integration, err := s.Orig.NewIntegration(l)
	if err != nil {
		return nil, err
	}

	id, err := s.Identifier(g)
	if err != nil {
		return nil, err
	}

	// Map from the original CommonConfig to the new settings. This is a 1:1
	// mapping, minus the loss of WALTruncateFrequency.
	origCommon := s.Orig.CommonConfig()

	if !origCommon.Enabled {
		return nil, fmt.Errorf("disabled integrations cannot be used in integrations-next")
	}

	newCommon := CommonConfig{
		InstanceKey:          origCommon.InstanceKey,
		ScrapeIntegration:    origCommon.ScrapeIntegration,
		ScrapeInterval:       origCommon.ScrapeInterval,
		ScrapeTimeout:        origCommon.ScrapeTimeout,
		RelabelConfigs:       origCommon.RelabelConfigs,
		MetricRelabelConfigs: origCommon.MetricRelabelConfigs,
	}

	// Generate our handler. Original integrations didn't accept a prefix, and
	// just assumed that they would be wired to /metrics somewhere.
	handler, err := v1Integration.MetricsHandler()
	if err != nil {
		return nil, fmt.Errorf("generating http handler: %w", err)
	} else if handler == nil {
		handler = http.NotFoundHandler()
	}

	// Generate targets. Original integrations used a static set of targets,
	// so this mapping can always be generated just once.
	//
	// Targets are generated from the result of ScrapeConfigs(), which returns a
	// tuple of job name and relative metrics path.
	//
	// Job names were prefixed at the subsystem level with integrations/, so we
	// will retain that behavior here.
	v1ScrapeConfigs := v1Integration.ScrapeConfigs()
	targets := make([]handlerTarget, 0, len(v1ScrapeConfigs))
	for _, sc := range v1ScrapeConfigs {
		targets = append(targets, handlerTarget{
			MetricsPath: sc.MetricsPath,
			Labels: model.LabelSet{
				model.JobLabel: model.LabelValue("integrations/" + sc.JobName),
			},
		})
	}

	// Convert he run function. Original integrations sometimes returned
	// ctx.Err() on exit. This isn't recommended anymore, but we need to hide the
	// error if it happens, since the error was previously ignored.
	runFunc := func(ctx context.Context) error {
		err := v1Integration.Run(ctx)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled) && ctx.Err() != nil:
			// Hide error that no longer happens in newer integrations.
			return nil
		default:
			return err
		}
	}

	// Aggregate our converted settings into a v2 integration.
	return &metricsHandlerIntegration{
		integrationName: s.Orig.Name(),
		instanceID:      id,

		common:  newCommon,
		globals: g,
		handler: handler,
		targets: targets,

		runFunc: runFunc,
	}, nil
}
