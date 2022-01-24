package config

import (
	"testing"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/util/subset"
	"github.com/stretchr/testify/require"
	apiext_v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestIntegration(t *testing.T) {
	toJSON := func(in string) apiext_v1.JSON {
		t.Helper()
		out, err := yaml.YAMLToJSONStrict([]byte(in))
		require.NoError(t, err)
		return apiext_v1.JSON{Raw: out}
	}

	tt := []struct {
		name   string
		input  map[string]interface{}
		expect string
	}{
		{
			name: "configured integration",
			input: map[string]interface{}{
				"agent": &gragent.GrafanaAgent{},
				"integration": &gragent.MetricsIntegration{
					Spec: gragent.MetricsIntegrationSpec{
						Name: "mysqld_exporter",
						Type: gragent.IntegrationTypeNormal,
						Config: toJSON(`
              data_source_names: root@(server-a:3306)/
            `),
					},
				},
			},
			expect: `
      autoscrape:
        enable: false
      data_source_names: root@(server-a:3306)/
      `,
		},
		{
			name: "integration no config",
			input: map[string]interface{}{
				"agent": &gragent.GrafanaAgent{},
				"integration": &gragent.MetricsIntegration{
					Spec: gragent.MetricsIntegrationSpec{
						Name: "mysqld_exporter",
						Type: gragent.IntegrationTypeNormal,
					},
				},
			},
			expect: `
      autoscrape:
        enable: false
      `,
		},
		{
			name: "extra_labels",
			input: map[string]interface{}{
				"agent": &gragent.GrafanaAgent{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-grafanaagent",
						Namespace: "monitoring",
					},
				},
				"integration": &gragent.MetricsIntegration{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-integration",
						Namespace: "default",
						Labels: map[string]string{
							"label-a": "label-a-value",
							"label-b": "label-b-value",
						},
					},
					Spec: gragent.MetricsIntegrationSpec{
						Name: "mysqld_exporter",
						Type: gragent.IntegrationTypeNormal,
					},
				},
			},
			expect: `
      extra_labels:
        __meta_agentoperator_grafanaagent_name: some-grafanaagent
        __meta_agentoperator_grafanaagent_namespace: monitoring
        __meta_agentoperator_integration_type: normal
        __meta_agentoperator_integration_cr_name: some-integration
        __meta_agentoperator_integration_cr_namespace: default
        __meta_agentoperator_integration_cr_label_label_a: label-a-value
        __meta_agentoperator_integration_cr_label_label_b: label-b-value
        __meta_agentoperator_integration_cr_labelpresent_label_a: "true"
        __meta_agentoperator_integration_cr_labelpresent_label_b: "true"
      `,
		},
		{
			name: "extra_labels merge",
			input: map[string]interface{}{
				"agent": &gragent.GrafanaAgent{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-grafanaagent",
						Namespace: "monitoring",
					},
				},
				"integration": &gragent.MetricsIntegration{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-integration",
						Namespace: "default",
					},
					Spec: gragent.MetricsIntegrationSpec{
						Name: "mysqld_exporter",
						Type: gragent.IntegrationTypeNormal,
						Config: toJSON(`
              extra_labels: 
                hello: world
            `),
					},
				},
			},
			expect: `
      extra_labels:
        # Make sure that our custom label exists with at least some of our
        # custom labels
        __meta_agentoperator_integration_cr_name: some-integration
        __meta_agentoperator_integration_cr_namespace: default
        hello: world
      `,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			vm, err := createVM(testStore())
			require.NoError(t, err)

			actual, err := runSnippetTLA(t, vm, "./integration.libsonnet", tc.input)
			require.NoError(t, err)
			require.NoError(t, subset.YAMLAssert([]byte(tc.expect), []byte(actual)), "incomplete yaml\n%s", actual)
		})
	}
}
