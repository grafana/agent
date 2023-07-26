package kubelet

import (
	"testing"

	"github.com/prometheus/prometheus/discovery/targetgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	bearer_token_file = "/path/to/file.token"
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")

	// Make sure that URL defaults to https://localhost:10250
	var args2 Arguments
	err = river.Unmarshal([]byte{}, &args2)
	require.NoError(t, err)
	require.Equal(t, args2.URL.String(), "https://localhost:10250")
}

func TestPodDeletion(t *testing.T) {
	pod1 := newPod("pod-1", "namespace-1")
	pod2 := newPod("pod-2", "namespace-2")
	podList1 := v1.PodList{
		Items: []v1.Pod{pod1, pod2},
	}
	podList2 := v1.PodList{
		Items: []v1.Pod{pod2},
	}

	kubeletDiscovery, err := NewKubeletDiscovery(DefaultConfig)
	require.NoError(t, err)

	_, err = kubeletDiscovery.refresh(podList1)
	require.NoError(t, err)
	require.Len(t, kubeletDiscovery.discoveredPodSources, 2)

	tg2, err := kubeletDiscovery.refresh(podList2)
	require.NoError(t, err)
	require.Len(t, kubeletDiscovery.discoveredPodSources, 1)
	// should contain a target group for pod 1 with an empty target list as it has
	// been deleted
	require.Contains(t, tg2, &targetgroup.Group{Source: podSource(pod1)})
}

func newPod(name, namespace string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "container-1",
					Ports: []v1.ContainerPort{
						{
							Name:          "port-1",
							ContainerPort: 443,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			PodIP: "1.2.3.4",
		},
	}
}
