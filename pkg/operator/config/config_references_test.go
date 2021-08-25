package config

import (
	"testing"

	"github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeployment_AssetReferences(t *testing.T) {
	deployment := Deployment{
		Agent: &v1alpha1.GrafanaAgent{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "agent",
			},
			Spec: v1alpha1.GrafanaAgentSpec{
				APIServerConfig: &prom.APIServerConfig{
					BasicAuth: &prom.BasicAuth{
						Username: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "spec-apiserverconfig-basicauth-username",
							},
							Key: "key",
						},
					},
				},
			},
		},
		Metrics: []MetricsInstance{{
			Instance: &v1alpha1.MetricsInstance{
				ObjectMeta: v1.ObjectMeta{Namespace: "metrics-instance"},
			},
			PodMonitors: []*prom.PodMonitor{{
				ObjectMeta: v1.ObjectMeta{Namespace: "pmon"},
			}},
			Probes: []*prom.Probe{{
				ObjectMeta: v1.ObjectMeta{Namespace: "probe"},
			}},
			ServiceMonitors: []*prom.ServiceMonitor{{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "smon",
				},
				Spec: prom.ServiceMonitorSpec{
					Endpoints: []prom.Endpoint{{
						BearerTokenSecret: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "prometheis-servicemonitors-spec-endpoints-bearertokensecret",
							},
							Key: "key",
						},
					}},
				},
			}},
		}},
	}

	require.Equal(t, []AssetReference{
		{
			Namespace: "agent",
			Reference: prom.SecretOrConfigMap{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "spec-apiserverconfig-basicauth-username",
					},
					Key: "key",
				},
			},
		},
		{
			Namespace: "smon",
			Reference: prom.SecretOrConfigMap{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "prometheis-servicemonitors-spec-endpoints-bearertokensecret",
					},
					Key: "key",
				},
			},
		},
	}, deployment.AssetReferences())
}
