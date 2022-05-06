package testcomponents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "testcomponents.tick",
		Config:  TickConfig{},
		Exports: TickExports{},

		Build: func(o component.Options, c component.Config) (component.Component, error) {
			return NewTick(o, c.(TickConfig))
		},
	})
}

// TickConfig configures the testcomponents.tick component.
type TickConfig struct {
	Frequency time.Duration `hcl:"frequency,attr"`
}

type TickExports struct {
	Time time.Time `hcl:"tick_time,optional"`
}

// Tick implements the testcomponents.tick component, where the wallclock time
// will be emitted on a given frequency.
type Tick struct {
	opts component.Options
	log  log.Logger

	cfgMut sync.Mutex
	cfg    TickConfig
}

func NewTick(o component.Options, cfg TickConfig) (*Tick, error) {
	t := &Tick{opts: o, log: o.Logger}
	if err := t.Update(cfg); err != nil {
		return nil, err
	}
	return t, nil
}

var (
	_ component.Component = (*Tick)(nil)
)

// Run implements Component.
func (t *Tick) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(t.getNextTick()):
			level.Info(t.log).Log("msg", "ticked")
			t.opts.OnStateChange(TickExports{Time: time.Now()})
		}
	}
}

func (t *Tick) getNextTick() time.Duration {
	t.cfgMut.Lock()
	defer t.cfgMut.Unlock()
	return t.cfg.Frequency
}

// Run implements Component.
func (t *Tick) Update(newConfig component.Config) error {
	t.cfgMut.Lock()
	defer t.cfgMut.Unlock()

	cfg := newConfig.(TickConfig)
	if cfg.Frequency == 0 {
		return fmt.Errorf("frequency must not be 0")
	}

	level.Info(t.log).Log("msg", "setting tick frequency", "freq", cfg.Frequency)
	return nil
}
