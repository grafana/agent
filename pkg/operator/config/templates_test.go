package config

import (
	"fmt"
	"strings"
	"testing"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/util"
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExternalLabels(t *testing.T) {
	tt := []struct {
		name   string
		input  interface{}
		expect string
	}{
		{
			name: "defaults",
			input: Deployment{
				Agent: &v1alpha1.GrafanaAgent{
					ObjectMeta: meta_v1.ObjectMeta{
						Namespace: "operator",
						Name:      "agent",
					},
				},
			},
			expect: util.Untab(`
				cluster: operator/agent
				__replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
			`),
		},
		{
			name: "external_labels",
			input: Deployment{
				Agent: &v1alpha1.GrafanaAgent{
					ObjectMeta: meta_v1.ObjectMeta{
						Namespace: "operator",
						Name:      "agent",
					},
					Spec: v1alpha1.GrafanaAgentSpec{
						Prometheus: v1alpha1.PrometheusSubsystemSpec{
							ExternalLabels: map[string]string{"foo": "bar"},
						},
					},
				},
			},
			expect: util.Untab(`
				cluster: operator/agent
				foo: bar
				__replica__: replica-$(STATEFULSET_ORDINAL_NUMBER)
			`),
		},
		{
			name: "custom labels",
			input: Deployment{
				Agent: &v1alpha1.GrafanaAgent{
					ObjectMeta: meta_v1.ObjectMeta{
						Namespace: "operator",
						Name:      "agent",
					},
					Spec: v1alpha1.GrafanaAgentSpec{
						Prometheus: v1alpha1.PrometheusSubsystemSpec{
							PrometheusExternalLabelName: strPointer("deployment"),
							ReplicaExternalLabelName:    strPointer("replica"),
							ExternalLabels:              map[string]string{"foo": "bar"},
						},
					},
				},
			},
			expect: util.Untab(`
				deployment: operator/agent
				foo: bar
				replica: replica-$(STATEFULSET_ORDINAL_NUMBER)
			`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			vm, err := createVM(nil)
			require.NoError(t, err)
			bb, err := jsonnetMarshal(tc.input)
			require.NoError(t, err)

			vm.TLACode("ctx", string(bb))
			actual, err := runSnippet(vm, "./component/external_labels.libsonnet", "ctx")
			require.NoError(t, err)
			require.YAMLEq(t, tc.expect, actual)
		})
	}
}

func TestKubeSDConfig(t *testing.T) {
	store := make(assets.SecretStore)

	store[assets.KeyForConfigMap("operator", &v1.ConfigMapKeySelector{
		LocalObjectReference: v1.LocalObjectReference{Name: "obj"},
		Key:                  "key",
	})] = "secretcm"

	store[assets.KeyForSecret("operator", &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{Name: "obj"},
		Key:                  "key",
	})] = "secretkey"

	tt := []struct {
		name   string
		input  map[string]interface{}
		expect string
	}{
		{
			name: "defaults",
			input: map[string]interface{}{
				"namespace": "operator",
				"role":      "pod",
			},
			expect: util.Untab(`
				role: pod
			`),
		},
		{
			name: "defaults",
			input: map[string]interface{}{
				"namespace":  "operator",
				"namespaces": []string{"operator"},
				"role":       "pod",
			},
			expect: util.Untab(`
				role: pod
				namespaces:
					names: [operator]
			`),
		},
		{
			name: "host",
			input: map[string]interface{}{
				"namespace": "operator",
				"apiServer": &prom_v1.APIServerConfig{Host: "host"},
				"role":      "pod",
			},
			expect: util.Untab(`
				role: pod
				api_server: host
			`),
		},
		{
			name: "basic auth",
			input: map[string]interface{}{
				"namespace": "operator",
				"apiServer": &prom_v1.APIServerConfig{
					BasicAuth: &prom_v1.BasicAuth{
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
				"role": "pod",
			},
			expect: util.Untab(`
				role: pod
				basic_auth:
					username: secretkey
					password: secretkey
			`),
		},
		{
			name: "bearer auth",
			input: map[string]interface{}{
				"namespace": "operator",
				"apiServer": &prom_v1.APIServerConfig{
					BearerToken:     "bearer",
					BearerTokenFile: "file",
				},
				"role": "pod",
			},
			expect: util.Untab(`
				role: pod
				authorization:
					type: Bearer
					credentials: bearer
					credentials_file: file
			`),
		},
		{
			name: "tls_config",
			input: map[string]interface{}{
				"namespace": "operator",
				"apiServer": &prom_v1.APIServerConfig{
					TLSConfig: &prom_v1.TLSConfig{
						CAFile: "ca",
					},
				},
				"role": "pod",
			},
			expect: util.Untab(`
				role: pod
				tls_config:
					ca_file: ca
			`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			vm, err := createVM(store)
			require.NoError(t, err)

			args := []string{"namespace", "namespaces", "apiServer", "role"}
			for _, arg := range args {
				bb, err := jsonnetMarshal(tc.input[arg])
				require.NoError(t, err)
				vm.TLACode(arg, string(bb))
			}

			actual, err := runSnippet(vm, "./component/kube_sd_config.libsonnet", args...)
			require.NoError(t, err)
			require.YAMLEq(t, tc.expect, actual)
		})
	}
}

func runSnippet(vm *jsonnet.VM, filename string, args ...string) (string, error) {
	boundArgs := make([]string, len(args))
	for i := range args {
		boundArgs[i] = fmt.Sprintf("%[1]s=%[1]s", args[i])
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
