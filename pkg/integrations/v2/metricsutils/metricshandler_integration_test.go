package metricsutils

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandlerIntegration_Targets(t *testing.T) {
	globals := integrations.Globals{
		AgentIdentifier: "testagent",
		AgentBaseURL: func() *url.URL {
			u, err := url.Parse("http://testagent/")
			require.NoError(t, err)
			return u
		}(),
		SubsystemOpts: integrations.DefaultSubsystemOptions,
	}

	t.Run("Extra labels", func(t *testing.T) {
		common := common.MetricsConfig{
			ExtraLabels: labels.FromMap(map[string]string{"foo": "bar", "fizz": "buzz"}),
		}
		common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)

		i, err := NewMetricsHandlerIntegration(nil, fakeConfig{}, common, globals, http.NotFoundHandler())
		require.NoError(t, err)

		actual := i.Targets(integrations.Endpoint{Host: "test", Prefix: "/test/"})
		expect := []*targetgroup.Group{{
			Source: "fake/testagent",
			Labels: model.LabelSet{
				"instance":       "testagent",
				"job":            "integrations/fake",
				"agent_hostname": "testagent",

				"__meta_agent_integration_name":       "fake",
				"__meta_agent_integration_instance":   "testagent",
				"__meta_agent_integration_autoscrape": "1",

				"foo":  "bar",
				"fizz": "buzz",
			},
			Targets: []model.LabelSet{{
				"__address__":      "test", // from integrations.Endpoint
				"__metrics_path__": "/test/metrics",
			}},
		}}
		require.Equal(t, expect, actual)
	})
}

type fakeConfig struct{}

func (fakeConfig) Name() string                                      { return "fake" }
func (fakeConfig) ApplyDefaults(_ integrations.Globals) error        { return nil }
func (fakeConfig) Identifier(g integrations.Globals) (string, error) { return g.AgentIdentifier, nil }
func (fakeConfig) NewIntegration(_ log.Logger, _ integrations.Globals) (integrations.Integration, error) {
	return nil, fmt.Errorf("not implemented")
}
