// Package kubernetes implements the logic for remote.kubernetes.secret and remote.kubernetes.configmap component.
package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/kubernetes"
	"github.com/grafana/river/rivertypes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_go "k8s.io/client-go/kubernetes"
)

type ResourceType string

const (
	TypeSecret    ResourceType = "secret"
	TypeConfigMap ResourceType = "configmap"
)

// Arguments control the component.
type Arguments struct {
	Namespace     string        `river:"namespace,attr"`
	Name          string        `river:"name,attr"`
	PollFrequency time.Duration `river:"poll_frequency,attr,optional"`
	PollTimeout   time.Duration `river:"poll_timeout,attr,optional"`

	// Client settings to connect to Kubernetes.
	Client kubernetes.ClientArguments `river:"client,block,optional"`
}

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	PollFrequency: 1 * time.Minute,
	PollTimeout:   15 * time.Second,
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
		return fmt.Errorf("poll_timeout must not be greater than 0")
	}
	return nil
}

// Exports holds settings exported by this component.
type Exports struct {
	Data map[string]rivertypes.OptionalSecret `river:"data,attr"`
}

// Component implements the remote.kubernetes.* component.
type Component struct {
	log  log.Logger
	opts component.Options

	mut  sync.Mutex
	args Arguments

	client *client_go.Clientset
	kind   ResourceType

	lastPoll    time.Time
	lastExports Exports // Used for determining whether exports should be updated

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New returns a new, unstarted remote.kubernetes.* component.
func New(opts component.Options, args Arguments, rType ResourceType) (*Component, error) {
	c := &Component{
		log:  opts.Logger,
		opts: opts,

		kind: rType,
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

// Run starts the remote.kubernetes.* component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.nextPoll()):
			c.poll()
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
			Message:    "got " + string(c.kind),
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

	data := map[string]rivertypes.OptionalSecret{}
	if c.kind == TypeSecret {
		secret, err := c.client.CoreV1().Secrets(c.args.Namespace).Get(ctx, c.args.Name, v1.GetOptions{})
		if err != nil {
			return err
		}
		for k, v := range secret.Data {
			data[k] = rivertypes.OptionalSecret{
				Value:    string(v),
				IsSecret: true,
			}
		}
	} else if c.kind == TypeConfigMap {
		cmap, err := c.client.CoreV1().ConfigMaps(c.args.Namespace).Get(ctx, c.args.Name, v1.GetOptions{})
		if err != nil {
			return err
		}
		for k, v := range cmap.Data {
			data[k] = rivertypes.OptionalSecret{
				Value:    v,
				IsSecret: false,
			}
		}
	}

	newExports := Exports{
		Data: data,
	}

	// Only send a state change event if the exports have changed from the
	// previous poll.
	if !reflect.DeepEqual(newExports.Data, c.lastExports.Data) {
		c.opts.OnStateChange(newExports)
	}

	c.lastExports = newExports
	return nil
}

// Update updates the remote.kubernetes.* component. After the update completes, a
// poll is forced.
func (c *Component) Update(args component.Arguments) (err error) {
	// defer initial poll so the lock is released first
	defer func() {
		if err != nil {
			return
		}
		// Poll after updating and propagate the error if the poll fails. If an error
		// occurred during Update, we don't bother to do anything.
		// It is important to set err and the health so startup works correctly
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

	return err
}

// CurrentHealth returns the current health of the component.
func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()
	return c.health
}
