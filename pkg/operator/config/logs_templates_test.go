package config

import (
	"fmt"
	"strings"
	"testing"

	jsonnet "github.com/google/go-jsonnet"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/util"
)

func TestLogsClientConfig(t *testing.T) {
	tt := []struct {
		name   string
		input  map[string]interface{}
		expect string
	}{
		{
			name: "all-in-one URL",
			input: map[string]interface{}{
				"namespace": "operator",
				"spec": &gragent.LogsClientSpec{
					URL: "http://username:password@localhost:3100/loki/api/v1/push",
				},
			},
			expect: util.Untab(`
				url: http://username:password@localhost:3100/loki/api/v1/push
			`),
		},
		{
			name: "full basic config",
			input: map[string]interface{}{
				"namespace": "operator",
				"spec": &gragent.LogsClientSpec{
					URL:       "http://localhost:3100/loki/api/v1/push",
					TenantID:  "tenant",
					BatchWait: "5m",
					BatchSize: 500,
					Timeout:   "5m",
					ExternalLabels: map[string]string{
						"foo":  "bar",
						"fizz": "buzz",
					},
					ProxyURL: "http://proxy:3100/",
					BackoffConfig: &gragent.LogsBackoffConfigSpec{
						MinPeriod:  "500ms",
						MaxPeriod:  "5m",
						MaxRetries: 100,
					},
				},
			},
			expect: util.Untab(`
				url: http://localhost:3100/loki/api/v1/push
				tenant_id: tenant
				batchwait: 5m
				batchsize: 500
				proxy_url: http://proxy:3100/
				backoff_config:
					min_period: 500ms
					max_period: 5m
					max_retries: 100
				external_labels:
					foo: bar
					fizz: buzz
				timeout: 5m
			`),
		},
		{
			name: "tls config",
			input: map[string]interface{}{
				"namespace": "operator",
				"spec": &gragent.LogsClientSpec{
					URL: "http://localhost:3100/loki/api/v1/push",
					TLSConfig: &prom.TLSConfig{
						CAFile:   "ca",
						KeyFile:  "key",
						CertFile: "cert",
					},
				},
			},
			expect: util.Untab(`
				url: http://localhost:3100/loki/api/v1/push
				tls_config:
					ca_file: ca
					key_file: key
					cert_file: cert
			`),
		},
		{
			name: "bearer tokens",
			input: map[string]interface{}{
				"namespace": "operator",
				"spec": &gragent.LogsClientSpec{
					URL:             "http://localhost:3100/loki/api/v1/push",
					BearerToken:     "tok",
					BearerTokenFile: "tokfile",
				},
			},
			expect: util.Untab(`
				url: http://localhost:3100/loki/api/v1/push
				bearer_token: tok
				bearer_token_file: tokfile
			`),
		},
		{
			name: "basic auth",
			input: map[string]interface{}{
				"namespace": "operator",
				"spec": &gragent.LogsClientSpec{
					URL: "http://localhost:3100/loki/api/v1/push",
					BasicAuth: &prom.BasicAuth{
						Username: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: "obj"},
							Key:                  "key",
						},
						Password: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: "obj"},
							Key:                  "key",
						},
					},
				},
			},
			expect: util.Untab(`
				url: http://localhost:3100/loki/api/v1/push
				basic_auth:
					username: secretkey
					password: secretkey
			`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			vm, err := createVM(testStore())
			require.NoError(t, err)

			actual, err := runSnippetTLA(t, vm, "./component/logs/client.libsonnet", tc.input)
			require.NoError(t, err)
			require.YAMLEq(t, tc.expect, actual)
		})
	}
}

func runSnippetTLA(t *testing.T, vm *jsonnet.VM, filename string, tla map[string]interface{}) (string, error) {
	t.Helper()

	args := make([]string, 0, len(tla))
	for arg := range tla {
		args = append(args, arg)
	}

	boundArgs := make([]string, len(args))
	for i := range args {
		boundArgs[i] = fmt.Sprintf("%[1]s=%[1]s", args[i])
	}

	// Bind argument to TLA.
	for arg, value := range tla {
		bb, err := jsonnetMarshal(value)
		require.NoError(t, err)
		vm.TLACode(arg, string(bb))
	}

	return vm.EvaluateAnonymousSnippet(
		filename,
		fmt.Sprintf(`
			local marshal = import './ext/marshal.libsonnet';
			local optionals = import './ext/optionals.libsonnet';
			local eval = import '%s';
			function(%s) marshal.YAML(optionals.trim(eval(%s)))
		`, filename, strings.Join(args, ","), strings.Join(boundArgs, ",")),
	)
}
