package operator

import (
	"testing"

	"github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_generateMetricsStatefulSetSpec(t *testing.T) {
	var (
		cfg   = &Config{}
		name  = "example"
		shard = int32(1)
	)

	t.Run("image should have version", func(t *testing.T) {
		h := grafana_v1alpha1.Hierarchy{
			Agent: &v1alpha1.GrafanaAgent{
				ObjectMeta: v1.ObjectMeta{Name: name, Namespace: name},
			},
		}

		spec, err := generateMetricsStatefulSetSpec(cfg, name, h, shard)
		require.NoError(t, err)
		require.Equal(t, DefaultAgentImage, spec.Template.Spec.Containers[1].Image)
	})

	t.Run("allow custom version", func(t *testing.T) {
		h := grafana_v1alpha1.Hierarchy{
			Agent: &v1alpha1.GrafanaAgent{
				ObjectMeta: v1.ObjectMeta{Name: name, Namespace: name},
				Spec: v1alpha1.GrafanaAgentSpec{
					Version: "vX.Y.Z",
				},
			},
		}

		spec, err := generateMetricsStatefulSetSpec(cfg, name, h, shard)
		require.NoError(t, err)
		require.Equal(t, DefaultAgentBaseImage+":vX.Y.Z", spec.Template.Spec.Containers[1].Image)
	})
}
