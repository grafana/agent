package config

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s_yaml "sigs.k8s.io/yaml"

	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
)

func TestBuildConfig(t *testing.T) {
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
					prometheus:
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
					http_listen_port: 8080
					log_level: debug

				prometheus:
					wal_directory: /var/lib/grafana-agent/data
					global:
						scrape_interval: 15s
						scrape_timeout: 10s
						external_labels:
							cluster: prod
							foo: bar
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
						prometheus:
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
										key: key
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
						http_listen_port: 8080
						log_level: debug

					prometheus:
						wal_directory: /var/lib/grafana-agent/data
						global:
							scrape_interval: 15s
							scrape_timeout: 10s
							external_labels:
								cluster: prod
								foo: bar
							remote_write:
							- url: http://localhost:9090/api/v1/write
								basic_auth:
									username: somesecret
									password: somesecret
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
			result, err := d.BuildConfig(store)
			require.NoError(t, err)

			if !assert.YAMLEq(t, tc.expect, result) {
				fmt.Println(result)
			}
		})
	}
}
