package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "remote.http",
		BuildComponent: func(l log.Logger, c Config) (component.Component[Config], error) {
			return NewComponent(l, c)
		},
	})
}

type Config struct {
	URL     string `hcl:"url" cty:"url"`
	Refresh string `hcl:"refresh" cty:"refresh"`
}

type State struct {
	Content string `hcl:"content" cty:"content"`
}

type Component struct {
	cfgMut      sync.Mutex
	cfg         Config
	refreshRate time.Duration
	updated     chan struct{}

	mut   sync.RWMutex
	state State

	log log.Logger
}

func NewComponent(l log.Logger, cfg Config) (*Component, error) {
	c := &Component{
		log:     l,
		updated: make(chan struct{}, 1),
	}
	if err := c.Update(cfg); err != nil {
		return nil, err
	}
	return c, nil
}

var _ component.Component[Config] = (*Component)(nil)

func (c *Component) Run(ctx context.Context, onStateChange func()) error {
	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	if err := c.refresh(onStateChange); err != nil {
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

			if err := c.refresh(onStateChange); err != nil {
				level.Error(c.log).Log("msg", "failed to get key from http", "err", err)
				// TODO(rfratto): set health?
			}
		}
	}
}

func (c *Component) refresh(onStateChange func()) error {
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
		onStateChange()
	}

	return nil
}

func (c *Component) Update(cfg Config) error {
	c.cfgMut.Lock()
	defer c.cfgMut.Unlock()

	spew.Dump(cfg)

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

func (c *Component) CurrentState() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.state
}

func (c *Component) Config() Config {
	return Config{}
}
