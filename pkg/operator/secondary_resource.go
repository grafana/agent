package operator

import (
	"github.com/go-kit/kit/log"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// secondaryResource is a secondary resource that is consumed by the primary
// resource (the GrafanaAgent CR) that should trigger a reconcile when it
// changes.
type secondaryResource int

// List of secondary resources that the reconciler will watch.
const (
	resourcePromInstance secondaryResource = iota
	resourceServiceMonitor
	resourcePodMonitor
	resourceProbe
	resourceSecret
)

// secondaryResources is the list of valid secondaryResources.
var secondaryResources = []secondaryResource{
	resourcePromInstance,
	resourceServiceMonitor,
	resourcePodMonitor,
	resourceProbe,
	resourceSecret,
}

// eventHandlers is a set of EnqueueRequestForSelector event handlers, one per
// secondary resource.
type eventHandlers map[secondaryResource]*enqueueRequestForSelector

// newResourceEventHandlers creates a new eventHandlers for all secondary
// resources using the given client and logger.
func newResourceEventHandlers(c client.Reader, l log.Logger) eventHandlers {
	m := make(eventHandlers)
	for _, r := range secondaryResources {
		m[r] = &enqueueRequestForSelector{Client: c, Log: l}
	}
	return m
}

// Clear informs all event handlers to stop sending events for the given
// namespaced name.
func (ev eventHandlers) Clear(name types.NamespacedName) {
	for _, v := range ev {
		v.Notify(name, nil)
	}
}
