package aws_firehose

import (
	"context"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/util"
	"sync"
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

type Component struct {
	fanout        []loki.LogsReceiver
	serverMetrics *util.UncheckedCollector
	handler       loki.LogsReceiver
	opts          component.Options
	mut           sync.RWMutex
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:          o,
		handler:       make(loki.LogsReceiver),
		fanout:        args.ForwardTo,
		serverMetrics: util.NewUncheckedCollector(nil),
	}

	o.Registerer.MustRegister(c.serverMetrics)

	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	defer func() {
		level.Info(c.opts.Logger).Log("msg", "loki.source.awsfirehose component shutting down, stopping the targets")
		c.mut.RLock()
		// todo(pablo): uncomment once target is hooked up
		//err := c.target.Stop()
		//if err != nil {
		//	level.Error(c.opts.Logger).Log("msg", "error while stopping gcplog target", "err", err)
		//}
		c.mut.RUnlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler:
			c.mut.RLock()
			for _, receiver := range c.fanout {
				receiver <- entry
			}
			c.mut.RUnlock()
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	//TODO implement me
	panic("implement me")
}
