// These tests depend on test assets from controller-runtime which don't work on Windows.

// +build !windows

package operator

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hashicorp/go-getter"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var (
	envTestK8sVersion = "1.19.2"

	envtestToolsURL = fmt.Sprintf(
		"https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-%s-%s-%s.tar.gz",
		envTestK8sVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
)

// TestEnqueueRequestForSelector creates an example Kubenretes cluster and runs
// EnqueueRequestForSelector to validate it works.
func TestEnqueueRequestForSelector(t *testing.T) {
	l := log.NewNopLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var env envtest.Environment
	setupEnvtest(t, &env)

	cfg, err := env.Start()
	require.NoError(t, err)
	t.Cleanup(func() { _ = env.Stop() })

	cli, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	// Tests will rely on a namespace existing, so let's create a namespace with
	// some labels.
	testNs := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "enqueue-test",
			Labels: map[string]string{"foo": "bar"},
		},
	}
	err = cli.Create(ctx, &testNs)
	require.NoError(t, err)

	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:      "test-pod",
		Namespace: "enqueue-test",
		Labels:    map[string]string{"fizz": "buzz"},
	}}

	t.Run("no watchers", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Create(event.CreateEvent{Object: testPod}, q)

		require.Equal(t, 0, q.Len(), "no changes should have been enqueued")
	})

	t.Run("matches watcher", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, []ResourceSelector{{
			NamespaceName:   NamespaceSelector{Any: true},
			NamespaceLabels: parseSelector(t, "foo in (bar)"),
			Labels:          parseSelector(t, "fizz in (buzz)"),
		}})

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 1, q.Len(), "expected one enqueue")
	})

	t.Run("matches watcher with explicit namespace", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, []ResourceSelector{{
			NamespaceName:   NamespaceSelector{MatchNames: []string{"enqueue-test"}},
			NamespaceLabels: labels.Everything(),
			Labels:          labels.Everything(),
		}})

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 1, q.Len(), "expected one enqueue")
	})

	t.Run("bad namespace name selector", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, []ResourceSelector{{
			NamespaceName:   NamespaceSelector{MatchNames: []string{"default"}},
			NamespaceLabels: labels.Everything(),
			Labels:          labels.Everything(),
		}})

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 0, q.Len(), "expected no enqueues")
	})

	t.Run("bad namespace label selector", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, []ResourceSelector{{
			NamespaceName:   NamespaceSelector{Any: true},
			NamespaceLabels: parseSelector(t, "foo notin (bar)"),
			Labels:          parseSelector(t, "fizz in (buzz)"),
		}})

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 0, q.Len(), "no changes should have been enqueued")
	})

	t.Run("bad label selector", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, []ResourceSelector{{
			NamespaceName:   NamespaceSelector{Any: true},
			NamespaceLabels: parseSelector(t, "foo in (bar)"),
			Labels:          parseSelector(t, "fizz notin (buzz)"),
		}})

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 0, q.Len(), "no changes should have been enqueued")
	})
}

// setupEnvtest downloads Envtest dependencies to a temporary directory.
func setupEnvtest(t *testing.T, env *envtest.Environment) {
	t.Helper()
	storagePath := t.TempDir()

	err := getter.Get(storagePath, envtestToolsURL)
	require.NoError(t, err, "failed to download dependencies for envtest")

	env.BinaryAssetsDirectory = filepath.Join(storagePath, "kubebuilder", "bin")
}

func parseSelector(t *testing.T, selector string) labels.Selector {
	t.Helper()
	s, err := labels.Parse(selector)
	require.NoError(t, err)
	return s
}
