// Package instances implements the grafana.cloud.instances component.
package instances

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/grafana/cloud"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/river/rivertypes"
)

func init() {
	component.Register(component.Registration{
		Name:    "grafana.cloud.instances",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments control the grafana.cloud.instances component.
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

// Exports holds settings exported by grafana.cloud.instances.
type Exports struct {
	Stacks map[string]Stack `river:"stacks,attr"`
}

type Stack struct {
	MetricsUsername  int    `river:"metrics_username,attr"`
	MetricsUrl       string `river:"metrics_url,attr"`
	LogsUsername     int    `river:"logs_username,attr"`
	LogsUrl          string `river:"logs_url,attr"`
	TracesUsername   int    `river:"traces_username,attr"`
	TracesUrl        string `river:"traces_url,attr"`
	ProfilesUsername int    `river:"profiles_username,attr"`
	ProfilesUrl      string `river:"profiles_url,attr"`
	OtlpUsername     int    `river:"otlp_username,attr"`
	OtlpUrl          string `river:"otlp_url,attr"`
}

// Component implements the grafana.cloud.instances component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut         sync.Mutex
	args        Arguments
	lastPoll    time.Time
	lastExports Exports // Used for determining whether exports should be updated
	// cli         *http.Client

	// Updated is written to whenever args updates.
	updated chan struct{}

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New returns a new, unstarted, grafana.cloud.instances component.
func New(opts component.Options, args Arguments) (*Component, error) {
	c := &Component{
		log:  opts.Logger,
		opts: opts,

		updated: make(chan struct{}, 1),

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

// Run starts the grafana.cloud.instances component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.nextPoll()):
			c.poll()
		case <-c.updated:
			// no-op; force the next wait to be reread.
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

// poll performs a HTTP GET for the component's configured URL. c.mut must
// not be held when calling. After polling, the component's health is updated
// with the success or failure status.
func (c *Component) poll() {
	err := c.pollError()
	c.updatePollHealth(err)
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
	c.mut.Lock()
	defer c.mut.Unlock()

	c.lastPoll = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), c.args.PollTimeout)
	defer cancel()

	lsr, err := cloud.GetListStacks(string(c.args.Token), c.args.Org, ctx)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to retrieve list of stacks from grafana cloud", "err", err)
		return fmt.Errorf("failed to retrieve list of stacks from grafana cloud: %w", err)
	}

	newExports := Exports{Stacks: make(map[string]Stack, len(lsr.Items))}
	for _, stack := range lsr.Items {
		newExports.Stacks[stack.Slug] = Stack{
			MetricsUsername:  stack.MetricsUsername,
			MetricsUrl:       stack.MetricsUrl,
			LogsUsername:     stack.LogsUsername,
			LogsUrl:          stack.LogsUrl,
			TracesUsername:   stack.TracesUsername,
			TracesUrl:        stack.TracesUrl,
			ProfilesUsername: stack.ProfilesUsername,
			ProfilesUrl:      stack.ProfilesUrl,
			OtlpUsername:     stack.OtlpUsername,
			OtlpUrl:          stack.OtlpUrl,
		}

	}

	// Only send a state change event if the exports have changed from the
	// previous poll.
	if len(c.lastExports.Stacks) != len(newExports.Stacks) {
		c.opts.OnStateChange(newExports)
	}

	for key, value1 := range c.lastExports.Stacks {
		if value2, ok := newExports.Stacks[key]; !ok || value1 != value2 {
			c.opts.OnStateChange(newExports)
		}
	}

	c.lastExports = newExports
	return nil
}

// Update updates the grafana.cloud.instances component. After the update completes, a
// poll is forced.
func (c *Component) Update(args component.Arguments) (err error) {
	// Poll after updating and propagate the error if the poll fails. If an error
	// occurred during Update, we don't bother to do anything.
	//
	// It's important to propagate the error in update so the initial state of
	// the component is calculated correctly, otherwise the exports will be empty
	// and may cause unexpected errors in downstream components.
	defer func() {
		if err != nil {
			return
		}
		err = c.pollError()
		c.updatePollHealth(err)
	}()

	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.args = newArgs

	// Override default UserAgent if another is provided in "headers" section
	// cli, err := prom_config.NewClientFromConfig(
	// 	*newArgs.Client.Convert(),
	// 	c.opts.ID,
	// 	prom_config.WithUserAgent(customUserAgent),
	// )
	// if err != nil {
	// 	return err
	// }
	// c.cli = cli

	// Send an updated event if one wasn't already read.
	select {
	case c.updated <- struct{}{}:
	default:
	}
	return nil
}

// CurrentHealth returns the current health of the component.
func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()
	return c.health
}
