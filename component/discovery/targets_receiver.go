package discovery

import (
	"reflect"
	"sync"
)

type TargetsSubscriber func([]Target)

// TargetsReceiver allows components to publish and subscribe to []Target updates.
type TargetsReceiver interface {
	// Publish publishes a new set of targets from a given source. The source is used to identify the publisher, so that
	// it can request to be removed via RemovePublisher.
	Publish(source string, message []Target)
	// RemovePublisher removes the publisher with the given identifier. This will remove targets that were most recently
	// published by that publisher and, if necessary, publish a new set of targets to all subscribers.
	RemovePublisher(identifier string)

	// Subscribe subscribes to updates of all the current targets from all publishers. The callback will be called with
	// the concatenated set of the most recent targets from all active publishers.
	Subscribe(identifier string, callback TargetsSubscriber)
	// Unsubscribe removes the subscriber with the given identifier, so that it will no longer receive updates when
	// targets are updated.
	Unsubscribe(identifier string)
}

type concatenatingTargetsReceiver struct {
	lock           sync.Mutex
	currentTargets map[string][]Target
	subscribers    map[string]TargetsSubscriber
}

var _ TargetsReceiver = (*concatenatingTargetsReceiver)(nil)

func NewConcatenatingTargetsReceiver() TargetsReceiver {
	return &concatenatingTargetsReceiver{
		currentTargets: make(map[string][]Target),
		subscribers:    make(map[string]TargetsSubscriber),
	}
}

func (c *concatenatingTargetsReceiver) RemovePublisher(identifier string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Remove the publisher if it exists
	prevTargets, prevExists := c.currentTargets[identifier]
	if !prevExists {
		return
	}
	delete(c.currentTargets, identifier)

	// If effective set of targets changes, publish it.
	if len(prevTargets) == 0 {
		return
	}
	c.concatAndPublishAll()
}

func (c *concatenatingTargetsReceiver) Subscribe(identifier string, callback TargetsSubscriber) {
	if callback == nil {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.subscribers[identifier] = callback
}

func (c *concatenatingTargetsReceiver) Unsubscribe(identifier string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.subscribers, identifier)
}

func (c *concatenatingTargetsReceiver) Publish(source string, newTargets []Target) {
	c.lock.Lock()
	defer c.lock.Unlock()

	previous, exists := c.currentTargets[source]
	if exists && reflect.DeepEqual(previous, newTargets) {
		// No change, nothing to publish
		return
	}
	c.currentTargets[source] = newTargets

	c.concatAndPublishAll()
}

// concatAndPublishAll concatenates all targets and publishes them to all subscribers. The lock must be held when
// calling this function.
func (c *concatenatingTargetsReceiver) concatAndPublishAll() {
	// Concatenate all targets into a single slice.
	concatenatedTargets := make([]Target, 0, len(c.currentTargets))
	for _, targets := range c.currentTargets {
		concatenatedTargets = append(concatenatedTargets, targets...)
	}

	// Publish the concatenated targets to all subscribers.
	for _, callback := range c.subscribers {
		callback(concatenatedTargets)
	}
}

type TargetsReceivers []TargetsReceiver

func (t *TargetsReceivers) Empty() bool {
	return len(*t) == 0
}

func (t *TargetsReceivers) Publish(source string, message []Target) {
	for _, tr := range *t {
		tr.Publish(source, message)
	}
}

func (t *TargetsReceivers) RemovePublisher(identifier string) {
	for _, tr := range *t {
		tr.RemovePublisher(identifier)
	}
}

func (t *TargetsReceivers) Subscribe(identifier string, callback TargetsSubscriber) {
	for _, tr := range *t {
		tr.Subscribe(identifier, callback)
	}
}

func (t *TargetsReceivers) Unsubscribe(identifier string) {
	for _, tr := range *t {
		tr.Unsubscribe(identifier)
	}
}
