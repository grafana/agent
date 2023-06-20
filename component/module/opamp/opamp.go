// Package git implements the module.git component.
package git

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module"
	"github.com/open-telemetry/opamp-go/protobufs"
	"google.golang.org/protobuf/proto"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.opamp",
		Args:    Arguments{},
		Exports: module.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments configures the module.git component.
type Arguments struct {
	URL           string            `river:"url,attr"`
	Labels        map[string]string `river:"labels,attr,optional"`
	PullFrequency time.Duration     `river:"pull_frequency,attr,optional"`

	Arguments map[string]any `river:"arguments,block,optional"`
}

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	PullFrequency: time.Minute,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Component implements the module.git component.
type Component struct {
	opts component.Options
	log  log.Logger
	mod  *module.ModuleComponent

	mut  sync.RWMutex
	args Arguments

	argsChanged chan struct{}

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
	_ component.HTTPComponent   = (*Component)(nil)
)

// New creates a new module.git component.
func New(o component.Options, args Arguments) (*Component, error) {
	m, err := module.NewModuleComponent(o)
	if err != nil {
		return nil, err
	}
	c := &Component{
		opts: o,
		log:  o.Logger,

		mod: m,

		argsChanged: make(chan struct{}, 1),
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go c.mod.RunFlowController(ctx)

	var (
		ticker  *time.Ticker
		tickerC <-chan time.Time
	)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-c.argsChanged:
			c.mut.Lock()
			{
				level.Info(c.log).Log("msg", "updating repository pull frequency", "new_frequency", c.args.PullFrequency)

				if c.args.PullFrequency > 0 {
					if ticker == nil {
						ticker = time.NewTicker(c.args.PullFrequency)
						tickerC = ticker.C
					} else {
						ticker.Reset(c.args.PullFrequency)
					}
				} else {
					if ticker != nil {
						ticker.Stop()
					}
					ticker = nil
					tickerC = nil
				}
			}
			c.mut.Unlock()

		case <-tickerC:
			level.Info(c.log).Log("msg", "updating repository", "new_frequency", c.args.PullFrequency)
			c.tickSendRequest(ctx)
		}
	}
}

func (c *Component) tickSendRequest(ctx context.Context) {
	c.mut.Lock()
	err := c.sendRequest(ctx, c.args)
	c.mut.Unlock()

	c.updateHealth(err)
}

func (c *Component) updateHealth(err error) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	if err != nil {
		c.health = component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    err.Error(),
			UpdateTime: time.Now(),
		}
	} else {
		c.health = component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "module updated",
			UpdateTime: time.Now(),
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) (err error) {
	defer func() {
		c.updateHealth(err)
	}()

	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)

	if err := c.sendRequest(context.Background(), newArgs); err != nil {
		return err
	}

	// Schedule an update for handling the changed arguments.
	select {
	case c.argsChanged <- struct{}{}:
	default:
	}

	c.args = newArgs
	return nil
}

// pollFile fetches the latest content from the repository and updates the
// controller. pollFile must only be called with c.mut held.
func (c *Component) sendRequest(ctx context.Context, args Arguments) error {
	// Prepare your proto request
	request := c.newAgentToServer(args)

	// Marshal the request to binary format
	body, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, args.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the Content-Type header to indicate Protocol Buffers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Scope-OrgID", "1")

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Unmarshal the response from binary format
	response := &protobufs.ServerToAgent{}
	if err := proto.Unmarshal(responseBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	cfgMap := response.RemoteConfig.GetConfig().GetConfigMap()
	cfg := cfgMap["remote_config"].GetBody()
	level.Info(c.log).Log("msg", "updating cfg", "cfg", string(cfg))

	return c.mod.LoadFlowContent(args.Arguments, string(cfg))
}

func (c *Component) newAgentToServer(args Arguments) *protobufs.AgentToServer {
	req := &protobufs.AgentToServer{AgentDescription: &protobufs.AgentDescription{IdentifyingAttributes: []*protobufs.KeyValue{}}}
	for key, val := range args.Labels {
		req.AgentDescription.IdentifyingAttributes = append(req.AgentDescription.IdentifyingAttributes, &protobufs.KeyValue{
			Key:   key,
			Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: val}},
		})
	}

	req.InstanceUid = c.opts.ID

	health := c.CurrentHealth()
	req.Health = &protobufs.AgentHealth{
		Healthy:   health.Health == component.HealthTypeHealthy,
		LastError: health.Message,
	}
	return req
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()

	return component.LeastHealthy(c.health, c.mod.CurrentHealth())
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	return c.mod.HTTPHandler()
}
