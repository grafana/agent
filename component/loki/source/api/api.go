package api

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/api/internal/lokipush"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.api",
		Args: Arguments{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Server               *fnet.ServerConfig  `river:",squash"`
	ForwardTo            []loki.LogsReceiver `river:"forward_to,attr"`
	Labels               map[string]string   `river:"labels,attr,optional"`
	RelabelRules         relabel.Rules       `river:"relabel_rules,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = Arguments{
		Server: fnet.DefaultServerConfig(),
	}
}

func (a *Arguments) labelSet() model.LabelSet {
	labelSet := make(model.LabelSet, len(a.Labels))
	for k, v := range a.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}
	return labelSet
}

type Component struct {
	opts               component.Options
	entriesChan        chan loki.Entry
	uncheckedCollector *util.UncheckedCollector

	serverMut sync.Mutex
	server    *lokipush.PushAPIServer

	// Use separate receivers mutex to address potential deadlock when Update drains the current server.
	// e.g. https://github.com/grafana/agent/issues/3391
	receiversMut sync.RWMutex
	receivers    []loki.LogsReceiver
}

func New(opts component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:               opts,
		entriesChan:        make(chan loki.Entry),
		receivers:          args.ForwardTo,
		uncheckedCollector: util.NewUncheckedCollector(nil),
	}
	opts.Registerer.MustRegister(c.uncheckedCollector)
	err := c.Update(args)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Component) Run(ctx context.Context) (err error) {
	defer c.stop()

	for {
		select {
		case entry := <-c.entriesChan:
			c.receiversMut.RLock()
			receivers := c.receivers
			c.receiversMut.RUnlock()

			for _, receiver := range receivers {
				select {
				case receiver.Chan() <- entry:
				case <-ctx.Done():
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	newArgs, ok := args.(Arguments)
	if !ok {
		return fmt.Errorf("invalid type of arguments: %T", args)
	}

	// if no server config provided, we'll use defaults
	if newArgs.Server == nil {
		newArgs.Server = &fnet.ServerConfig{}
	}
	// to avoid port conflicts, if no GRPC is configured, make sure we use a random port
	// also, use localhost IP, so we don't require root to run.
	if newArgs.Server.GRPC == nil {
		newArgs.Server.GRPC = &fnet.GRPCConfig{
			ListenPort:    0,
			ListenAddress: "127.0.0.1",
		}
	}

	c.receiversMut.Lock()
	c.receivers = newArgs.ForwardTo
	c.receiversMut.Unlock()

	c.serverMut.Lock()
	defer c.serverMut.Unlock()
	serverNeedsRestarting := c.server == nil || !reflect.DeepEqual(c.server.ServerConfig(), *newArgs.Server)
	if serverNeedsRestarting {
		if c.server != nil {
			c.server.Shutdown()
		}

		// [server.Server] registers new metrics every time it is created. To
		// avoid issues with re-registering metrics with the same name, we create a
		// new registry for the server every time we create one, and pass it to an
		// unchecked collector to bypass uniqueness checking.
		serverRegistry := prometheus.NewRegistry()
		c.uncheckedCollector.SetCollector(serverRegistry)

		var err error
		c.server, err = lokipush.NewPushAPIServer(c.opts.Logger, newArgs.Server, loki.NewEntryHandler(c.entriesChan, func() {}), serverRegistry)
		if err != nil {
			return fmt.Errorf("failed to create embedded server: %v", err)
		}
		err = c.server.Run()
		if err != nil {
			return fmt.Errorf("failed to run embedded server: %v", err)
		}
	}

	c.server.SetLabels(newArgs.labelSet())
	c.server.SetRelabelRules(newArgs.RelabelRules)
	c.server.SetKeepTimestamp(newArgs.UseIncomingTimestamp)

	return nil
}

func (c *Component) stop() {
	c.serverMut.Lock()
	defer c.serverMut.Unlock()
	if c.server != nil {
		c.server.Shutdown()
		c.server = nil
	}
}
