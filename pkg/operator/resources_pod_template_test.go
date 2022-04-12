package operator

import (
	"testing"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_generatePodTemplate(t *testing.T) {
	var (
		cfg  = &Config{}
		name = "example"
	)

	t.Run("image should have version", func(t *testing.T) {
		deploy := gragent.Deployment{
			Agent: &gragent.GrafanaAgent{
				ObjectMeta: v1.ObjectMeta{Name: name, Namespace: name},
			},
		}

		tmpl, _, err := generatePodTemplate(cfg, "agent", deploy, podTemplateOptions{})
		require.NoError(t, err)
		require.Equal(t, DefaultAgentImage, tmpl.Spec.Containers[1].Image)
	})

	t.Run("allow custom version", func(t *testing.T) {
		deploy := gragent.Deployment{
			Agent: &gragent.GrafanaAgent{
				ObjectMeta: v1.ObjectMeta{Name: name, Namespace: name},
				Spec: gragent.GrafanaAgentSpec{
					Version: "vX.Y.Z",
				},
			},
		}

		tmpl, _, err := generatePodTemplate(cfg, "agent", deploy, podTemplateOptions{})
		require.NoError(t, err)
		require.Equal(t, DefaultAgentBaseImage+":vX.Y.Z", tmpl.Spec.Containers[1].Image)
	})
}
