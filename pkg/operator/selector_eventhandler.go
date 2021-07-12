package operator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	promop_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// enqueueRequestForSelector allows for requesting that specific
// reconciliations occur whenever an object that matches a selector
// comes in.
//
// Implements handler.EventHandler.
type enqueueRequestForSelector struct {
	Client client.Reader
	Log    log.Logger

	mut      sync.RWMutex
	watchers map[types.NamespacedName][]ResourceSelector
}

// Create implements handler.EventHandler.
func (e *enqueueRequestForSelector) Create(ev event.CreateEvent, q workqueue.RateLimitingInterface) {
	level.Debug(e.logger(ev.Object)).Log("msg", "got create for object")
	e.handleEvent(ev.Object, q)
}

// Update implements handler.EventHandler.
func (e *enqueueRequestForSelector) Update(ev event.UpdateEvent, q workqueue.RateLimitingInterface) {
	level.Debug(e.logger(ev.ObjectNew)).Log("msg", "got update for object")
	e.handleEvent(ev.ObjectOld, q)
	e.handleEvent(ev.ObjectNew, q)
}

// Delete implements handler.EventHandler.
func (e *enqueueRequestForSelector) Delete(ev event.DeleteEvent, q workqueue.RateLimitingInterface) {
	level.Debug(e.logger(ev.Object)).Log("msg", "got delete for object")
	e.handleEvent(ev.Object, q)
}

// Generic implements handler.EventHandler.
func (e *enqueueRequestForSelector) Generic(ev event.GenericEvent, q workqueue.RateLimitingInterface) {
	level.Debug(e.logger(ev.Object)).Log("msg", "got generic event for object")
	e.handleEvent(ev.Object, q)
}

func (e *enqueueRequestForSelector) logger(obj client.Object) log.Logger {
	gvk := obj.GetObjectKind().GroupVersionKind()
	return log.With(
		e.Log,
		"kind", fmt.Sprintf("%s/%s.%s", gvk.Group, gvk.Version, gvk.Kind),
		"key", client.ObjectKeyFromObject(obj),
	)
}

func (e *enqueueRequestForSelector) handleEvent(obj client.Object, q workqueue.RateLimitingInterface) {
	e.mut.RLock()
	defer e.mut.RUnlock()

	if e.watchers == nil {
		return
	}

	// Go through our watchers. If any of their selectors match this object,
	// enqueue a reconcile request for the watcher.
	for watcher, selectors := range e.watchers {
		performReconcile := false

		for _, selector := range selectors {
			if !e.namespaceNameMatches(obj.GetNamespace(), selector.NamespaceName) ||
				!e.namespaceMatches(obj.GetNamespace(), selector.NamespaceLabels) ||
				!selector.Labels.Matches(labels.Set(obj.GetLabels())) {
				continue
			}

			performReconcile = true
			break
		}

		if performReconcile {
			q.Add(reconcile.Request{NamespacedName: watcher})
		}
	}
}

// namespaceNameMatches checks to see if the namespace "in" matches the name
// selector provided.
func (e *enqueueRequestForSelector) namespaceNameMatches(in string, selector promop_v1.NamespaceSelector) bool {
	if selector.Any {
		return true
	}

	for _, n := range selector.MatchNames {
		if n == in {
			return true
		}
	}

	return false
}

// namespaceMatches checks to see if the namespace "in" matches the labels
// selector provided.
func (e *enqueueRequestForSelector) namespaceMatches(in string, selector labels.Selector) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var ns v1.Namespace
	if err := e.Client.Get(ctx, types.NamespacedName{Name: in}, &ns); err != nil {
		level.Error(e.Log).Log("msg", "failed to look up namespace", "namespace", in, "err", err)
		return false
	}

	return selector.Matches(labels.Set(ns.Labels))
}

// Notify will notify obj to reconcile if an event was received that matches
// any selector in ss.
//
// To stop being notified for changes, call Notify again with nil for ss.
func (e *enqueueRequestForSelector) Notify(obj types.NamespacedName, ss []ResourceSelector) {
	e.mut.Lock()
	defer e.mut.Unlock()

	if e.watchers == nil {
		e.watchers = make(map[types.NamespacedName][]ResourceSelector)
	}

	if ss == nil {
		delete(e.watchers, obj)
	} else {
		e.watchers[obj] = ss
	}
}

// NamespaceSelector re-exports the Prometheus Operator NamespaceSelector.
type NamespaceSelector = promop_v1.NamespaceSelector

// ResourceSelector is a combination of namespace and label selectors
// to match against an incoming resource.
type ResourceSelector struct {
	// NamespaceName matches the name of the namespace.
	NamespaceName NamespaceSelector

	// NamespaceLabels matches the labels on the namespace of the modified
	// resource.
	NamespaceLabels labels.Selector

	// Labels matches the labels on the modified resource.
	Labels labels.Selector
}
