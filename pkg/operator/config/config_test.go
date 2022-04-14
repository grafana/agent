package config

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_yaml "sigs.k8s.io/yaml"

	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
)

func TestBuildConfigMetrics(t *testing.T) {
	var store = make(assets.SecretStore)

	store[assets.Key("/secrets/default/example-secret/key")] = "somesecret"
	store[assets.Key("/configMaps/default/example-cm/key")] = "somecm"

	tt := []struct {
		input  string
		expect string
	}{
		{
			input: util.Untab(`
				metadata:
					name: example
					namespace: default
				spec:
					logLevel: debug
					metrics:
						scrapeInterval: 15s
						scrapeTimeout: 10s
						externalLabels:
							cluster: prod
							foo: bar
						remoteWrite:
						- name: rw-1
							url: http://localhost:9090/api/v1/write
			`),
			expect: util.Untab(`
				server:
					log_level: debug

				metrics:
					wal_directory: /var/lib/grafana-agent/data
					global:
						scrape_interval: 15s
						scrape_timeout: 10s
						external_labels:
							cluster: prod
							foo: bar
							__replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
						remote_write:
						- name: rw-1
							url: http://localhost:9090/api/v1/write
			`),
		},
		{
			input: util.Untab(`
					metadata:
						name: example
						namespace: default
					spec:
						logLevel: debug
						metrics:
							scrapeInterval: 15s
							scrapeTimeout: 10s
							externalLabels:
								cluster: prod
								foo: bar
							remoteWrite:
							- url: http://localhost:9090/api/v1/write
								basicAuth:
									username:
										name: example-secret
										key: key
									password:
										name: example-secret
										key: pword
								tlsConfig:
									ca:
										configMap:
											name:	example-cm
											key: key
									cert:
										secret:
											name: example-secret
											key: key
									keySecret:
										name: example-secret
										key: key
				`),
			expect: util.Untab(`
					server:
						log_level: debug

					metrics:
						wal_directory: /var/lib/grafana-agent/data
						global:
							scrape_interval: 15s
							scrape_timeout: 10s
							external_labels:
								cluster: prod
								foo: bar
								__replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
							remote_write:
							- url: http://localhost:9090/api/v1/write
								basic_auth:
									username: somesecret
									password_file: /var/lib/grafana-agent/secrets/_secrets_default_example_secret_pword
								tls_config:
									ca_file: /var/lib/grafana-agent/secrets/_configMaps_default_example_cm_key
									cert_file: /var/lib/grafana-agent/secrets/_secrets_default_example_secret_key
									key_file: /var/lib/grafana-agent/secrets/_secrets_default_example_secret_key
				`),
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("index_%d", i), func(t *testing.T) {
			var spec grafana.GrafanaAgent
			err := k8s_yaml.Unmarshal([]byte(tc.input), &spec)
			require.NoError(t, err)

			d := Deployment{Agent: &spec}
			result, err := d.BuildConfig(store, MetricsType)
			require.NoError(t, err)

			if !assert.YAMLEq(t, tc.expect, result) {
				fmt.Println(result)
			}
		})
	}
}

func TestAdditionalScrapeConfigsMetrics(t *testing.T) {
	var store = make(assets.SecretStore)

	additionalSelector := &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{Name: "configs"},
		Key:                  "configs",
	}

	input := Deployment{
		Agent: &grafana.GrafanaAgent{
			ObjectMeta: meta_v1.ObjectMeta{
				Namespace: "operator",
				Name:      "agent",
			},
			Spec: grafana.GrafanaAgentSpec{
				Image:              strPointer("grafana/agent:latest"),
				ServiceAccountName: "agent",
				Metrics: grafana.MetricsSubsystemSpec{
					InstanceSelector: &meta_v1.LabelSelector{
						MatchLabels: map[string]string{"agent": "agent"},
					},
				},
			},
		},
		Metrics: []MetricsInstance{{
			Instance: &grafana.MetricsInstance{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: "operator",
					Name:      "primary",
				},
				Spec: grafana.MetricsInstanceSpec{
					RemoteWrite: []grafana.RemoteWriteSpec{{
						URL: "http://cortex:80/api/prom/push",
					}},
					AdditionalScrapeConfigs: additionalSelector,
				},
			},
		}},
	}

	store[assets.KeyForSecret("operator", additionalSelector)] = util.Untab(`
	- job_name: job
		kubernetes_sd_configs:
		- role: node
	`)

	expect := util.Untab(`
server: {}

metrics:
  wal_directory: /var/lib/grafana-agent/data
  global:
    external_labels:
      __replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
      cluster: operator/agent
  configs:
  - name: operator/primary
    remote_write:
    - url: http://cortex:80/api/prom/push
    scrape_configs:
    - job_name: job
      kubernetes_sd_configs:
      - role: node
	`)

	result, err := input.BuildConfig(store, MetricsType)
	require.NoError(t, err)

	if !assert.YAMLEq(t, expect, result) {
		fmt.Println(result)
	}
}

func TestBuildConfigLogs(t *testing.T) {
	var store = make(assets.SecretStore)

	store[assets.Key("/secrets/default/example-secret/key")] = "somesecret"
	store[assets.Key("/configMaps/default/example-cm/key")] = "somecm"

	tt := []struct {
		input  string
		expect string
	}{
		{
			input: util.Untab(`
				metadata:
					name: example
					namespace: default
				spec:
					logLevel: debug
			`),
			expect: util.Untab(`
				server:
					log_level: debug
				logs:
					positions_directory: /var/lib/grafana-agent/data
			`),
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("index_%d", i), func(t *testing.T) {
			var spec grafana.GrafanaAgent
			err := k8s_yaml.Unmarshal([]byte(tc.input), &spec)
			require.NoError(t, err)

			d := Deployment{Agent: &spec}
			result, err := d.BuildConfig(store, LogsType)
			require.NoError(t, err)

			if !assert.YAMLEq(t, tc.expect, result) {
				fmt.Println(result)
			}
		})
	}
}

func strPointer(s string) *string { return &s }
