package shared

import (
	"testing"

	"github.com/go-kit/log"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/http"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/stretchr/testify/require"
)

//
// Tests for controller's utilization of the MetricsIntegration interface.
//

func Test_controller_MetricsIntegration_Targets(t *testing.T) {
	integrationWithTarget := func(targetName string) Integration {
		return mockMetricsIntegration{
			Integration: v2.NoOpIntegration,
			TargetsFunc: func(v2.Endpoint) []*targetgroup.Group {
				return []*targetgroup.Group{{
					Targets: []model.LabelSet{{model.AddressLabel: model.LabelValue(targetName)}},
				}}
			},
			ScrapeConfigsFunc: func(c discovery.Configs) []*autoscrape.ScrapeConfig { return nil },
		}
	}

	integrations := []Config{
		mockConfigNameTuple(t, "a", "instanceA").WithNewIntegrationFunc(func(l log.Logger, g v2.Globals) (Integration, error) {
			return integrationWithTarget("a"), nil
		}),
		mockConfigNameTuple(t, "b", "instanceB").WithNewIntegrationFunc(func(l log.Logger, g v2.Globals) (Integration, error) {
			return integrationWithTarget("b"), nil
		}),
	}

	t.Run("All", func(t *testing.T) {
		ctrl, err := v2.NewController(
			util.TestLogger(t),
			controllerConfig(integrations),
			v2.Globals{},
		)
		require.NoError(t, err)
		_ = v2.NewSyncController(t, ctrl)

		result := ctrl.Targets(v2.Endpoint{Prefix: "/"}, v2.TargetOptions{})
		expect := []*integrations.TargetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
			{Targets: []model.LabelSet{{model.AddressLabel: "b"}}},
		}
		require.Equal(t, expect, result)
	})

	t.Run("All by Integration", func(t *testing.T) {
		ctrl, err := v2.NewController(
			util.TestLogger(t),
			controllerConfig(integrations),
			v2.Globals{},
		)
		require.NoError(t, err)
		_ = v2.NewSyncController(t, ctrl)

		result := ctrl.Targets(v2.Endpoint{Prefix: "/"}, v2.TargetOptions{
			Integrations: []string{"a", "b"},
		})
		expect := []*integrations.TargetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
			{Targets: []model.LabelSet{{model.AddressLabel: "b"}}},
		}
		require.Equal(t, expect, result)
	})

	t.Run("Specific Integration", func(t *testing.T) {
		ctrl, err := v2.NewController(
			util.TestLogger(t),
			controllerConfig(integrations),
			v2.Globals{},
		)
		require.NoError(t, err)
		_ = v2.NewSyncController(t, ctrl)

		result := ctrl.Targets(v2.Endpoint{Prefix: "/"}, v2.TargetOptions{
			Integrations: []string{"a"},
		})
		expect := []*integrations.TargetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
		}
		require.Equal(t, expect, result)
	})
}

func Test_controller_MetricsIntegration_ScrapeConfig(t *testing.T) {
	integrationWithTarget := func(targetName string) Integration {
		return mockMetricsIntegration{
			Integration: v2.NoOpIntegration,
			ScrapeConfigsFunc: func(c discovery.Configs) []*autoscrape.ScrapeConfig {
				return []*autoscrape.ScrapeConfig{{
					Instance: "default",
					Config:   prom_config.ScrapeConfig{JobName: targetName},
				}}
			},
		}
	}

	integrations := []Config{
		mockConfigNameTuple(t, "a", "instanceA").WithNewIntegrationFunc(func(l log.Logger, g v2.Globals) (Integration, error) {
			return integrationWithTarget("a"), nil
		}),
		mockConfigNameTuple(t, "b", "instanceB").WithNewIntegrationFunc(func(l log.Logger, g v2.Globals) (Integration, error) {
			return integrationWithTarget("b"), nil
		}),
	}

	ctrl, err := v2.NewController(
		util.TestLogger(t),
		controllerConfig(integrations),
		v2.Globals{},
	)
	require.NoError(t, err)
	// NOTE(rfratto): we explicitly don't run the controller here because
	// ScrapeConfigs should return the list of scrape targets even when the
	// integration isn't running.

	result := ctrl.ScrapeConfigs("/", &http.DefaultSDConfig)
	expect := []*autoscrape.ScrapeConfig{
		{Instance: "default", Config: prom_config.ScrapeConfig{JobName: "a"}},
		{Instance: "default", Config: prom_config.ScrapeConfig{JobName: "b"}},
	}
	require.Equal(t, expect, result)
}

//
// Tests for controller's utilization of the MetricsIntegration interface.
//

type mockMetricsIntegration struct {
	Integration
	TargetsFunc       func(ep v2.Endpoint) []*targetgroup.Group
	ScrapeConfigsFunc func(discovery.Configs) []*autoscrape.ScrapeConfig
}

func (m mockMetricsIntegration) Targets(ep v2.Endpoint) []*targetgroup.Group {
	return m.TargetsFunc(ep)
}

func (m mockMetricsIntegration) ScrapeConfigs(cfgs discovery.Configs) []*autoscrape.ScrapeConfig {
	return m.ScrapeConfigsFunc(cfgs)
}
