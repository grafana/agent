package aws_firehose

import (
	"context"
	"reflect"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/aws_firehose/internal"
	"github.com/grafana/agent/pkg/util"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.awsfirehose",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Server       *fnet.ServerConfig  `river:",squash"`
	ForwardTo    []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

func (a *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*a = Arguments{}
	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}

	return nil
}

// Component is the main type for the `loki.source.awsfirehose` component.
type Component struct {
	// mut controls concurrent access to fanout
	mut    sync.RWMutex
	fanout []loki.LogsReceiver

	// destination is the main destination where the TargetServer writes received log entries to
	destination loki.LogsReceiver
	rbs         []*relabel.Config

	server *fnet.TargetServer

	opts component.Options
	args Arguments

	// utils
	serverMetrics *util.UncheckedCollector
	logger        log.Logger
}

// New creates a new Component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:          o,
		destination:   make(loki.LogsReceiver),
		fanout:        args.ForwardTo,
		serverMetrics: util.NewUncheckedCollector(nil),

		logger: log.With(o.Logger, "component", "aws_firehose_logs"),
	}

	o.Registerer.MustRegister(c.serverMetrics)

	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run starts a routine forwards received message to each destination component.
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()
		c.shutdownServer()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.destination:
			c.mut.RLock()
			for _, receiver := range c.fanout {
				receiver <- entry
			}
			c.mut.RUnlock()
		}
	}
}

// Update updates the component with a new configuration, restarting the server if needed.
func (c *Component) Update(args component.Arguments) error {
	var err error
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.fanout = newArgs.ForwardTo

	// todo(pablo): is it a good practice to keep a reference to the arguments in the
	// component struct, used for comparing here rather than destructuring them?
	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules) > 0 {
		c.rbs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules)
	}

	serverNeedsUpdate := !reflect.DeepEqual(c.args.Server, newArgs.Server)
	if !serverNeedsUpdate {
		c.args = newArgs
		return nil
	}

	c.shutdownServer()

	jobName := strings.Replace(c.opts.ID, ".", "_", -1)

	registry := prometheus.NewRegistry()
	c.serverMetrics.SetCollector(registry)

	wlog := log.With(c.logger, "component", "aws_firehose_logs")
	c.server, err = fnet.NewTargetServer(wlog, jobName, registry, newArgs.Server)
	if err != nil {
		return err
	}

	if err = c.server.MountAndRun(func(router *mux.Router) {
		// re-create handler when server is re-computed
		// todo(pablo): should use unchecked collector here?
		handler := internal.NewHandler(c, c.logger, c.opts.Registerer, c.rbs)
		router.Path("/awsfirehose/api/v1/push").Methods("POST").Handler(handler)
	}); err != nil {
		return err
	}

	c.args = newArgs
	return nil
}

// Send implements internal.Sender so that the component is able to receive logs decoded by the handler.
func (c *Component) Send(ctx context.Context, entry loki.Entry) {
	c.destination <- entry
}

// shutdownServer will shut down the currently used server.
// It is not goroutine-safe and mut write lock must be held when it's called.
func (c *Component) shutdownServer() {
	if c.server != nil {
		c.server.StopAndShutdown()
		c.server = nil
	}
}
