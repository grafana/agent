package v2

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	http_sd "github.com/prometheus/prometheus/discovery/http"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/stretchr/testify/require"
)

//
// Tests for controller's utilization of the MetricsIntegration interface.
//

func Test_controller_MetricsIntegration_Targets(t *testing.T) {
	integrationWithTarget := func(targetName string) Integration {
		return mockMetricsIntegration{
			Integration: NoOpIntegration,
			TargetsFunc: func(Endpoint) []*targetgroup.Group {
				return []*targetgroup.Group{{
					Targets: []model.LabelSet{{model.AddressLabel: model.LabelValue(targetName)}},
				}}
			},
			ScrapeConfigsFunc: func(c discovery.Configs) []*autoscrape.ScrapeConfig { return nil },
		}
	}

	integrations := []Config{
		mockConfigNameTuple(t, "a", "instanceA").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("a"), nil
		}),
		mockConfigNameTuple(t, "b", "instanceB").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("b"), nil
		}),
	}
	mockConfigs := &mockIntegrationConfigs{configs: integrations}

	t.Run("All", func(t *testing.T) {
		ctrl, err := NewController(
			util.TestLogger(t),
			mockConfigs,
			Globals{},
		)
		require.NoError(t, err)
		_ = NewSyncController(t, ctrl)

		result := ctrl.Targets(Endpoint{Prefix: "/"}, TargetOptions{})
		expect := []*shared.TargetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
			{Targets: []model.LabelSet{{model.AddressLabel: "b"}}},
		}
		require.Equal(t, expect, result)
	})

	t.Run("All by Integration", func(t *testing.T) {
		ctrl, err := NewController(
			util.TestLogger(t),
			mockConfigs,
			Globals{},
		)
		require.NoError(t, err)
		_ = NewSyncController(t, ctrl)

		result := ctrl.Targets(Endpoint{Prefix: "/"}, TargetOptions{
			Integrations: []string{"a", "b"},
		})
		expect := []*shared.TargetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
			{Targets: []model.LabelSet{{model.AddressLabel: "b"}}},
		}
		require.Equal(t, expect, result)
	})

	t.Run("Specific Integration", func(t *testing.T) {
		ctrl, err := NewController(
			util.TestLogger(t),
			mockConfigs,
			Globals{},
		)
		require.NoError(t, err)
		_ = NewSyncController(t, ctrl)

		result := ctrl.Targets(Endpoint{Prefix: "/"}, TargetOptions{
			Integrations: []string{"a"},
		})
		expect := []*shared.TargetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
		}
		require.Equal(t, expect, result)
	})
}

func Test_controller_MetricsIntegration_ScrapeConfig(t *testing.T) {
	integrationWithTarget := func(targetName string) Integration {
		return mockMetricsIntegration{
			Integration: NoOpIntegration,
			ScrapeConfigsFunc: func(c discovery.Configs) []*autoscrape.ScrapeConfig {
				return []*autoscrape.ScrapeConfig{{
					Instance: "default",
					Config:   prom_config.ScrapeConfig{JobName: targetName},
				}}
			},
		}
	}

	integrations := []Config{
		mockConfigNameTuple(t, "a", "instanceA").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("a"), nil
		}),
		mockConfigNameTuple(t, "b", "instanceB").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("b"), nil
		}),
	}

	mockConfigs := &mockIntegrationConfigs{configs: integrations}

	ctrl, err := NewController(
		util.TestLogger(t),
		mockConfigs,
		Globals{},
	)
	require.NoError(t, err)
	// NOTE(rfratto): we explicitly don't run the controller here because
	// ScrapeConfigs should return the list of scrape targets even when the
	// integration isn't running.

	result := ctrl.ScrapeConfigs("/", &http_sd.DefaultSDConfig)
	expect := []*autoscrape.ScrapeConfig{
		{Instance: "default", Config: prom_config.ScrapeConfig{JobName: "a"}},
		{Instance: "default", Config: prom_config.ScrapeConfig{JobName: "b"}},
	}
	require.Equal(t, expect, result)
}

//
// Tests for controller's utilization of the MetricsIntegration interface.
//
var (
	_ MetricsIntegration = (*mockMetricsIntegration)(nil)
)

type mockMetricsIntegration struct {
	Integration
	TargetsFunc       func(ep Endpoint) []*targetgroup.Group
	ScrapeConfigsFunc func(discovery.Configs) []*autoscrape.ScrapeConfig
}

func (m mockMetricsIntegration) Targets(ep Endpoint) []*targetgroup.Group {
	return m.TargetsFunc(ep)
}

func (m mockMetricsIntegration) ScrapeConfigs(configs discovery.Configs) []*autoscrape.ScrapeConfig {
	return m.ScrapeConfigsFunc(configs)
}
