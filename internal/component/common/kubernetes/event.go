package kubernetes

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/internal/flow/logging/level"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// This type must be hashable, so it is kept simple. The indexer will maintain a
// cache of current state, so this is mostly used for logging.
type Event struct {
	Typ       EventType
	ObjectKey string
}

type EventType string

const (
	EventTypeResourceChanged EventType = "resource-changed"
)

type queuedEventHandler struct {
	log   log.Logger
	queue workqueue.RateLimitingInterface
}

func NewQueuedEventHandler(log log.Logger, queue workqueue.RateLimitingInterface) *queuedEventHandler {
	return &queuedEventHandler{
		log:   log,
		queue: queue,
	}
}

// OnAdd implements the cache.ResourceEventHandler interface.
func (c *queuedEventHandler) OnAdd(obj interface{}, _ bool) {
	c.publishEvent(obj)
}

// OnUpdate implements the cache.ResourceEventHandler interface.
func (c *queuedEventHandler) OnUpdate(oldObj, newObj interface{}) {
	c.publishEvent(newObj)
}

// OnDelete implements the cache.ResourceEventHandler interface.
func (c *queuedEventHandler) OnDelete(obj interface{}) {
	c.publishEvent(obj)
}

func (c *queuedEventHandler) publishEvent(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to get key for object", "err", err)
		return
	}

	c.queue.AddRateLimited(Event{
		Typ:       EventTypeResourceChanged,
		ObjectKey: key,
	})
}
