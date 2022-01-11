//go:build !nonetwork && !nodocker
// +build !nonetwork,!nodocker

package operator

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util/k8s"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// TestEnqueueRequestForSelector creates an example Kubenretes cluster and runs
// EnqueueRequestForSelector to validate it works.
func TestEnqueueRequestForSelector(t *testing.T) {
	l := log.NewNopLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cluster, err := k8s.NewCluster(ctx, k8s.Options{})
	require.NoError(t, err)
	defer cluster.Stop()

	cli := cluster.Client()

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

		e.Notify(types.NamespacedName{Name: "watcher"}, buildSelectorSet(
			&namespaceLabelSelector{Selector: parseSelector(t, "foo in (bar)")},
			&labelSelector{Selector: parseSelector(t, "fizz in (buzz)")},
		))

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 1, q.Len(), "expected one enqueue")
	})

	t.Run("matches watcher with explicit namespace", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, buildSelectorSet(
			&namespaceSelector{Namespace: "enqueue-test"},
			&namespaceLabelSelector{Selector: labels.Everything()},
			&labelSelector{Selector: labels.Everything()},
		))

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 1, q.Len(), "expected one enqueue")
	})

	t.Run("bad namespace name selector", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, buildSelectorSet(
			&namespaceSelector{Namespace: "default"},
			&namespaceLabelSelector{Selector: labels.Everything()},
			&labelSelector{Selector: labels.Everything()},
		))

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 0, q.Len(), "expected no enqueues")
	})

	t.Run("bad namespace label selector", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, buildSelectorSet(
			&namespaceLabelSelector{Selector: parseSelector(t, "foo notin (bar)")},
			&labelSelector{Selector: parseSelector(t, "fizz in (buzz)")},
		))

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 0, q.Len(), "no changes should have been enqueued")
	})

	t.Run("bad label selector", func(t *testing.T) {
		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewRateLimitingQueue(limiter)

		e := enqueueRequestForSelector{Client: cli, Log: l}
		e.Notify(types.NamespacedName{Name: "watcher"}, buildSelectorSet(
			&namespaceLabelSelector{Selector: parseSelector(t, "foo in (bar)")},
			&labelSelector{Selector: parseSelector(t, "fizz notin (buzz)")},
		))

		e.Create(event.CreateEvent{Object: testPod}, q)
		require.Equal(t, 0, q.Len(), "no changes should have been enqueued")
	})
}

// buildSelectorSet returns a single multiSelector composed of ss.
func buildSelectorSet(ss ...resourceSelector) []resourceSelector {
	return []resourceSelector{&multiSelector{Selectors: ss}}
}

func parseSelector(t *testing.T, selector string) labels.Selector {
	t.Helper()
	s, err := labels.Parse(selector)
	require.NoError(t, err)
	return s
}
