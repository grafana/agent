package operator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/operator/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	watchers map[types.NamespacedName][]resourceSelector
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
		var performReconcile bool
		for _, selector := range selectors {
			if selector.Matches(e.Log, e.Client, obj) {
				performReconcile = true
				break
			}
		}
		if performReconcile {
			q.Add(reconcile.Request{NamespacedName: watcher})
		}
	}
}

// Notify will notify obj to reconcile if an event was received that matches
// any selector in ss.
//
// To stop being notified for changes, call Notify again with nil for ss.
func (e *enqueueRequestForSelector) Notify(obj types.NamespacedName, ss []resourceSelector) {
	e.mut.Lock()
	defer e.mut.Unlock()

	if e.watchers == nil {
		e.watchers = make(map[types.NamespacedName][]resourceSelector)
	}

	if ss == nil {
		delete(e.watchers, obj)
	} else {
		e.watchers[obj] = ss
	}
}

type resourceSelector interface {
	Matches(l log.Logger, c client.Reader, o client.Object) bool
	SetListOptions(lo *client.ListOptions)
}

// multiSelector returns true if all inner selectors match.
type multiSelector struct {
	Selectors []resourceSelector
}

func (s *multiSelector) Matches(l log.Logger, c client.Reader, o client.Object) bool {
	for _, inner := range s.Selectors {
		if !inner.Matches(l, c, o) {
			return false
		}
	}
	return true
}

func (s *multiSelector) SetListOptions(lo *client.ListOptions) {
	for _, inner := range s.Selectors {
		inner.SetListOptions(lo)
	}
}

type namespaceSelector struct {
	Namespace string
}

func (s *namespaceSelector) Matches(l log.Logger, c client.Reader, o client.Object) bool {
	return o.GetNamespace() == s.Namespace
}

func (s *namespaceSelector) SetListOptions(lo *client.ListOptions) {
	lo.Namespace = s.Namespace
}

type namespaceLabelSelector struct {
	Selector labels.Selector
}

func (s *namespaceLabelSelector) Matches(l log.Logger, c client.Reader, o client.Object) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	in := o.GetNamespace()
	var ns v1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: in}, &ns); err != nil {
		level.Error(l).Log("msg", "failed to look up namespace", "namespace", in, "err", err)
		return false
	}

	return s.Selector.Matches(labels.Set(ns.Labels))
}

func (s *namespaceLabelSelector) SetListOptions(lo *client.ListOptions) {
	// no-op
}

type labelSelector struct {
	Selector labels.Selector
}

func (s *labelSelector) Matches(l log.Logger, c client.Reader, o client.Object) bool {
	return s.Selector.Matches(labels.Set(o.GetLabels()))
}

func (s *labelSelector) SetListOptions(lo *client.ListOptions) {
	lo.LabelSelector = s.Selector
}

type assetReferenceSelector struct {
	Reference config.AssetReference
}

func (s *assetReferenceSelector) Matches(l log.Logger, c client.Reader, o client.Object) bool {
	if o.GetNamespace() != s.Reference.Namespace {
		return false
	}

	if sc, ok := o.(*v1.Secret); ok {
		return s.Reference.Reference.Secret != nil && sc.Name == s.Reference.Reference.Secret.Name
	} else if cm, ok := o.(*v1.ConfigMap); ok {
		return s.Reference.Reference.ConfigMap != nil && cm.Name == s.Reference.Reference.ConfigMap.Name
	}

	return false
}

func (s *assetReferenceSelector) SetListOptions(lo *client.ListOptions) {
	// no-op
}
