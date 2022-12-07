package operator

import (
	"testing"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/stretchr/testify/assert"
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

	t.Run("security ctx does not contain privileged", func(t *testing.T) {
		deploy := gragent.Deployment{
			Agent: &gragent.GrafanaAgent{
				ObjectMeta: v1.ObjectMeta{Name: name, Namespace: name},
			},
		}

		tmpl, _, err := generatePodTemplate(cfg, "agent", deploy, podTemplateOptions{})
		require.NoError(t, err)
		require.Equal(t, "config-reloader", tmpl.Spec.Containers[0].Name)
		assert.False(t, tmpl.Spec.Containers[0].SecurityContext.Privileged != nil &&
			*tmpl.Spec.Containers[0].SecurityContext.Privileged,
			"privileged is not required. Fargate cannot schedule privileged containers.")
	})
}

func TestSanitizeKubernetesLabel(t *testing.T) {
	validLabel := "this_1_is-a.valid_label"

	invalidCharLabel := "this#is$not&valid"
	invalidCharLabelExpected := "this.is.not.valid"

	invalidStartLabel := "-notvalid"
	invalidStartLabelExpected := "notvalid"

	require.Equal(t, validLabel, sanitizeKubernetesLabel(validLabel))
	require.Equal(t, invalidCharLabelExpected, sanitizeKubernetesLabel(invalidCharLabel))
	require.Equal(t, invalidStartLabelExpected, sanitizeKubernetesLabel(invalidStartLabel))
}
