package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:   "remote.http",
		Config: Config{},
		BuildComponent: func(o component.Options, c component.Config) (component.Component, error) {
			return NewComponent(o, c.(Config))
		},
	})
}

// Config represents the input state of the remote.http component.
type Config struct {
	URL     string `hcl:"url"`
	Refresh string `hcl:"refresh"`
}

// State represents the output state of the remote.http component.
type State struct {
	Content string `hcl:"content"`
}

// Component is the remote.http component.
type Component struct {
	log  log.Logger
	opts component.Options

	cfgMut      sync.Mutex
	cfg         Config
	refreshRate time.Duration
	updated     chan struct{}

	mut   sync.RWMutex
	state State
}

// NewComponent creates a new remote.http component.
func NewComponent(o component.Options, cfg Config) (*Component, error) {
	c := &Component{
		log:     o.Logger,
		opts:    o,
		updated: make(chan struct{}, 1),
	}
	if err := c.Update(cfg); err != nil {
		return nil, err
	}
	return c, nil
}

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	if err := c.refresh(); err != nil {
		level.Error(c.log).Log("msg", "failed to get key from http", "err", err)
		// TODO(rfratto): set health?
	}

	for {
		c.cfgMut.Lock()
		waitTime := c.refreshRate
		c.cfgMut.Unlock()

		select {
		case <-ctx.Done():
			return nil
		case <-c.updated:
			// no-op: go back to the start
		case <-time.After(waitTime):
			level.Debug(c.log).Log("msg", "refreshing key")

			if err := c.refresh(); err != nil {
				level.Error(c.log).Log("msg", "failed to get key from http", "err", err)
				// TODO(rfratto): set health?
			}
		}
	}
}

func (c *Component) refresh() error {
	resp, err := http.Get(c.cfg.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %s", resp.Status)
	}

	stringContent := strings.TrimSpace(string(bb))

	c.mut.Lock()
	defer c.mut.Unlock()

	if c.state.Content != stringContent {
		level.Info(c.log).Log("msg", "new value retrieved from http, emitting state updated message")
		c.state.Content = stringContent
		c.opts.OnStateChange()
	}

	return nil
}

// Update implements UpdatableComponent.
func (c *Component) Update(newConfig component.Config) error {
	cfg := newConfig.(Config)

	c.cfgMut.Lock()
	defer c.cfgMut.Unlock()

	refreshDuration, err := time.ParseDuration(cfg.Refresh)
	if err != nil {
		return err
	}

	c.refreshRate = refreshDuration
	c.cfg = cfg

	select {
	case c.updated <- struct{}{}:
	default:
	}
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.state
}

// Config implements Component.
func (c *Component) Config() Config {
	c.cfgMut.Lock()
	defer c.cfgMut.Unlock()
	return c.cfg
}
