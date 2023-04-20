package http

import (
	"context"
	"fmt"

	"github.com/efficientgo/core/errors"
	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/http/internal/lokipush"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
)

// TODO: this component also supports GRPC, so we may want to call it `loki.source.push_api` or something else.
const componentName = "loki.source.http"

type Arguments struct {
	HTTPAddress string `river:"http_address,attr"`
	HTTPPort    int    `river:"http_port,attr"`

	ForwardTo []loki.LogsReceiver `river:"forward_to,attr"`

	Labels               map[string]string `river:"labels,attr,optional"`
	RelabelRules         relabel.Rules     `river:"relabel_rules,attr,optional"`
	UseIncomingTimestamp bool              `river:"use_incoming_timestamp,attr,optional"`
	// TODO: allow to configure other Server fields in a dedicated block, to match promtail's
	//       https://grafana.com/docs/loki/next/clients/promtail/configuration/#server
}

type Component struct {
	opts component.Options
	args Arguments
}

func init() {
	component.Register(component.Registration{
		Name: componentName,
		Args: Arguments{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments)), nil
		},
	})
}

func New(opts component.Options, args Arguments) component.Component {
	return &Component{
		opts: opts,
		args: args,
	}
}

func (c *Component) Run(ctx context.Context) error {
	c.opts.Logger.Log("msg", "starting component")

	entriesChan := make(chan loki.Entry)
	entryHandler := loki.NewEntryHandler(
		entriesChan,
		func() {
			c.opts.Logger.Log("msg", "entry handler stopped")
		})

	pushTarget, err := lokipush.NewPushTarget(
		c.opts.Logger,
		entryHandler,
		relabel.ComponentToPromRelabelConfigs(c.args.RelabelRules),
		// TODO: pick correct metric names - avoid collisions!
		c.opts.ID,
		c.pushTargetConfig(),
	)

	if err != nil {
		return errors.Wrapf(err, "failed to create loki push API server: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			c.opts.Logger.Log("msg", "finishing due to context done")
			// When PushTarget is stopped, it will also Stop() the entry handler.
			return pushTarget.Stop()
		case entry := <-entriesChan:
			// TODO: fan out
			c.opts.Logger.Log("msg", fmt.Sprintf("got log message ====> %v", entry))
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	if newArgs, ok := args.(Arguments); !ok {
		return errors.Newf("invalid type of arguments: %T", args)
	} else {
		c.args = newArgs
	}
	// TODO: implement update properly...
	return nil
}

func (c *Component) pushTargetConfig() *lokipush.PushTargetConfig {
	return &lokipush.PushTargetConfig{
		Server: server.Config{
			HTTPListenPort:          c.args.HTTPPort,
			HTTPListenAddress:       c.args.HTTPAddress,
			Registerer:              c.opts.Registerer,
			MetricsNamespace:        "loki_source_http",
			RegisterInstrumentation: false,
			Log:                     logging.GoKit(log.With(c.opts.Logger, "component", componentName)),
		},
		Labels:        c.args.labelSet(),
		KeepTimestamp: c.args.UseIncomingTimestamp,
	}
}

func (a *Arguments) labelSet() model.LabelSet {
	labelSet := make(model.LabelSet, len(a.Labels))
	for k, v := range a.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}
	return labelSet
}
