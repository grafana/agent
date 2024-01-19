// Package receivers implements the grafana.cloud.receivers component.
package receivers

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/grafana/cloud"
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/river/rivertypes"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:    "grafana.cloud.receivers",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments control the grafana.cloud.receivers component.
type Arguments struct {
	Token         rivertypes.Secret `river:"token,attr"`
	Org           string            `river:"org,attr"`
	PollFrequency time.Duration     `river:"poll_frequency,attr,optional"`
	PollTimeout   time.Duration     `river:"poll_timeout,attr,optional"`
}

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	PollFrequency: 24 * time.Hour,
	PollTimeout:   20 * time.Second,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args.PollFrequency <= 0 {
		return fmt.Errorf("poll_frequency must be greater than 0")
	}
	if args.PollTimeout <= 0 {
		return fmt.Errorf("poll_timeout must be greater than 0")
	}
	if args.PollTimeout >= args.PollFrequency {
		return fmt.Errorf("poll_timeout must be less than poll_frequency")
	}

	return nil
}

// Exports holds settings exported by grafana.cloud.receivers.
type Exports struct {
	Stacks map[string]Stack `river:"stacks,attr"`
}

type Stack struct {
	PrometheusReceiver storage.Appendable `river:"prometheus_receiver,attr"`
}

// Component implements the grafana.cloud.receivers component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut      sync.Mutex
	args     Arguments
	lastPoll time.Time
	lastLsr  cloud.ListStacksResponse

	exportsMut sync.Mutex
	exports    Exports

	managedRemoteWrites map[string]*remotewrite.Component

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New returns a new, unstarted, grafana.cloud.receivers component.
func New(opts component.Options, args Arguments) (*Component, error) {
	c := &Component{
		log:  opts.Logger,
		opts: opts,

		exports: Exports{Stacks: make(map[string]Stack, 0)},

		managedRemoteWrites: make(map[string]*remotewrite.Component, 0),

		health: component.Health{
			Health:     component.HealthTypeUnknown,
			Message:    "component started",
			UpdateTime: time.Now(),
		},
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run starts the grafana.cloud.receivers component.
func (c *Component) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan error, 1)
	for _, rw := range c.managedRemoteWrites {
		go func(rw *remotewrite.Component) {
			err := rw.Run(ctx)
			if err != nil {
				ch <- err
			}
		}(rw)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.nextPoll()):
			c.mut.Lock()
			defer c.mut.Unlock()
			err := c.pollError()
			c.updatePollHealth(err)
			if err != nil {
				return err
			}
		case err := <-ch:
			return err
		}
	}
}

// nextPoll returns how long to wait to poll given the last time a
// poll occurred. nextPoll returns 0 if a poll should occur immediately.
func (c *Component) nextPoll() time.Duration {
	c.mut.Lock()
	defer c.mut.Unlock()

	nextPoll := c.lastPoll.Add(c.args.PollFrequency)
	now := time.Now()

	if now.After(nextPoll) {
		// Poll immediately; next poll period was in the past.
		return 0
	}
	return nextPoll.Sub(now)
}

func (c *Component) updatePollHealth(err error) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	if err == nil {
		c.health = component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "polled endpoint",
			UpdateTime: time.Now(),
		}
	} else {
		c.health = component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("polling failed: %s", err),
			UpdateTime: time.Now(),
		}
	}
}

// pollError is like poll but returns an error if one occurred.
func (c *Component) pollError() error {
	c.lastPoll = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), c.args.PollTimeout)
	defer cancel()

	lsr, err := cloud.GetListStacks(string(c.args.Token), c.args.Org, ctx)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to retrieve list of stacks from grafana cloud", "err", err)
		return fmt.Errorf("failed to retrieve list of stacks from grafana cloud: %w", err)
	}

	isUpdated := false
	if len(c.lastLsr.Items) != len(lsr.Items) {
		isUpdated = true
	} else {
		for i := 0; i < len(c.lastLsr.Items); i++ {
			if c.lastLsr.Items[i] != lsr.Items[i] {
				isUpdated = true
			}
		}
	}

	if isUpdated {
		c.lastLsr = lsr
		for _, stack := range lsr.Items {
			if _, ok := c.exports.Stacks[stack.Slug]; !ok {
				c.exports.Stacks[stack.Slug] = Stack{}
				c.managedRemoteWrites[stack.Slug], err = c.newManagedPrometheusRemoteWrite(stack.MetricsUrl, strconv.Itoa(stack.MetricsUsername), c.args.Token, stack.Slug)
				if err != nil {
					level.Error(c.log).Log("msg", "failed to initialize a prometheus receiver", "err", err)
					return fmt.Errorf("failed to initialize a prometheus receiver: %w", err)
				}
			}
		}
	}

	return nil
}

// Update updates the grafana.cloud.receivers component. After the update completes, a
// poll is forced.
func (c *Component) Update(args component.Arguments) (err error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.args = args.(Arguments)

	// Poll after updating and propagate the error if the poll fails.
	err = c.pollError()
	c.updatePollHealth(err)
	return err
}

// CurrentHealth returns the current health of the component.
func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()
	return c.health
}
