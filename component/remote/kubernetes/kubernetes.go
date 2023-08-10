// Package http implements the remote.http component.
package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_go "k8s.io/client-go/kubernetes"
)

func init() {
	component.Register(component.Registration{
		Name:    "remote.kubernetes.secret",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments control the remote.http component.
type Arguments struct {
	Namespace     string        `river:"namespace,attr"`
	Name          string        `river:"name,attr"`
	PollFrequency time.Duration `river:"poll_frequency,attr,optional"`

	// Client settings to connect to Kubernetes.
	Client kubernetes.ClientArguments `river:"client,block,optional"`
}

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	PollFrequency: 1 * time.Minute,
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
	return nil
}

// Exports holds settings exported by remote.http.
type Exports struct {
	Data map[string]string `river:"data,attr"`
}

// Component implements the remote.http component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut  sync.Mutex
	args Arguments

	client *client_go.Clientset

	lastPoll    time.Time
	lastExports Exports // Used for determining whether exports should be updated

	// Updated is written to whenever args updates.
	updated chan struct{}

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New returns a new, unstarted, remote.http component.
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

// Run starts the remote.http component.
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
			Message:    "got secret",
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

	// TODO: add timeout
	secret, err := c.client.CoreV1().Secrets(c.args.Namespace).Get(context.Background(), c.args.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	data := map[string]string{}
	for k, v := range secret.Data {
		data[k] = string(v)
	}
	newExports := Exports{
		Data: data,
	}

	// TODO: deep compare
	// Only send a state change event if the exports have changed from the
	// previous poll.
	//if c.lastExports != newExports {
	//	c.opts.OnStateChange(newExports)
	//}

	c.lastExports = newExports
	return nil
}

// Update updates the remote.http component. After the update completes, a
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

	restConfig, err := c.args.Client.BuildRESTConfig(c.log)
	if err != nil {
		return err
	}
	c.client, err = client_go.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}

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
