package discovery

import (
	"reflect"
	"sync"
)

type TargetsSubscriber func([]Target)

type TargetsReceiver interface {
	Publish(source string, message []Target)
	RemovePublisher(identifier string)

	Subscribe(identifier string, callback TargetsSubscriber)
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
	c.concatAndPublishAll(identifier)
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

	c.concatAndPublishAll(source)
}

// concatAndPublishAll concatenates all targets and publishes them to all subscribers. The lock must be held when
// calling this function.
func (c *concatenatingTargetsReceiver) concatAndPublishAll(source string) {
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

type TargetsReceiversProvider interface {
	TargetsReceivers() TargetsReceivers
}
