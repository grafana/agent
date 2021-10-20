package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/util"
)

func TestIntegrationConfig(t *testing.T) {
	tt := []struct {
		name   string
		input  map[string]interface{}
		expect string
	}{
		{
			name: "basic integration",
			input: map[string]interface{}{
				"instance": &gragent.IntegrationInstance{
					Spec: gragent.IntegrationInstanceSpec{
						Name:   "agent",
						Config: "",
					},
				},
			},
			expect: util.Untab(`
        agent:
          enabled: true
          scrape_integration: false
      `),
		},
		{
			name: "explicit enabled",
			input: map[string]interface{}{
				"instance": &gragent.IntegrationInstance{
					Spec: gragent.IntegrationInstanceSpec{
						Name:   "agent",
						Config: "enabled: true",
					},
				},
			},
			expect: util.Untab(`
        agent:
          enabled: true
          scrape_integration: false
      `),
		},
		{
			name: "explicit disabled",
			input: map[string]interface{}{
				"instance": &gragent.IntegrationInstance{
					Spec: gragent.IntegrationInstanceSpec{
						Name:   "agent",
						Config: "enabled: false",
					},
				},
			},
			expect: util.Untab(`
        agent:
          enabled: false
          scrape_integration: false
      `),
		},
		{
			name: "scrape_integration ignored",
			input: map[string]interface{}{
				"instance": &gragent.IntegrationInstance{
					Spec: gragent.IntegrationInstanceSpec{
						Name:   "agent",
						Config: "scrape_integration: true",
					},
				},
			},
			expect: util.Untab(`
        agent:
          enabled: true
          scrape_integration: false
      `),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			vm, err := createVM(testStore())
			require.NoError(t, err)

			actual, err := runSnippetTLA(t, vm, "./integration.libsonnet", tc.input)
			require.NoError(t, err)
			require.YAMLEq(t, tc.expect, actual)
		})
	}
}
