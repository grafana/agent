package lokipush

// This code is copied from Promtail. The fake package is used to configure
// fake client that can be used in testing.

import (
	"sync"

	"github.com/grafana/agent/component/common/loki"
)

// TODO(thampiotr): get rid of this file once https://github.com/grafana/agent/pull/3583 is merged

// FakeClient is a fake client used for testing.
type FakeClient struct {
	entries  loki.LogsReceiver
	received []loki.Entry
	once     sync.Once
	mtx      sync.Mutex
	wg       sync.WaitGroup
	OnStop   func()
}

func NewFakeClient(stop func()) *FakeClient {
	c := &FakeClient{
		OnStop:  stop,
		entries: make(loki.LogsReceiver),
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for e := range c.entries {
			c.mtx.Lock()
			c.received = append(c.received, e)
			c.mtx.Unlock()
		}
	}()
	return c
}

// Stop implements client.FakeClient
func (c *FakeClient) Stop() {
	c.once.Do(func() { close(c.entries) })
	c.wg.Wait()
	c.OnStop()
}

func (c *FakeClient) Chan() chan<- loki.Entry {
	return c.entries
}

func (c *FakeClient) Received() []loki.Entry {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	cpy := make([]loki.Entry, len(c.received))
	copy(cpy, c.received)
	return cpy
}

// StopNow implements client.FakeClient
func (c *FakeClient) StopNow() {
	c.Stop()
}

func (c *FakeClient) Name() string {
	return "fake"
}

// Clear is used to clean up the buffered received entries, so the same client can be re-used between
// test cases.
func (c *FakeClient) Clear() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.received = []loki.Entry{}
}
