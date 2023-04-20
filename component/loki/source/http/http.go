package http

import (
	"context"
	"fmt"
	"github.com/efficientgo/core/errors"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/http/internal/lokipush"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/server"
)

type Arguments struct {
	HttpAddress string              `river:"http_address,attr"`
	HttpPort    int                 `river:"http_port,attr"`
	ForwardTo   []loki.LogsReceiver `river:"forward_to,attr"`

	Labels               map[string]string `river:"labels,attr,optional"`
	RelabelRules         relabel.Rules     `river:"relabel_rules,attr,optional"`
	UseIncomingTimestamp bool              `river:"use_incoming_timestamp,attr,optional"`

	//TODO: add support for http_server_read_timeout like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
	//TODO: add support for http_server_write_timeout like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
	//TODO: add support for http_server_idle_timeout like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
}

type Component struct {
	opts component.Options
	args Arguments
}

func init() {
	component.Register(component.Registration{
		Name: "loki.source.http",
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

	entriesChan := make(chan api.Entry)
	entryHandler := api.NewEntryHandler(
		entriesChan,
		func() {
			c.opts.Logger.Log("msg", "entry handler stopped")
		})

	pushTarget, err := lokipush.NewPushTarget(
		c.opts.Logger,
		entryHandler,
		relabel.ComponentToPromRelabelConfigs(c.args.RelabelRules),
		// TODO: pick correct metric names
		"loki_source_http",
		c.args.toPushTargetConfig(),
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

func (a *Arguments) toPushTargetConfig() *scrapeconfig.PushTargetConfig {
	return &scrapeconfig.PushTargetConfig{
		Server: server.Config{
			HTTPListenPort:    a.HttpPort,
			HTTPListenAddress: a.HttpAddress,
		},
		Labels:        a.labelSet(),
		KeepTimestamp: a.UseIncomingTimestamp,
	}
}

func (a *Arguments) labelSet() model.LabelSet {
	labelSet := make(model.LabelSet, len(a.Labels))
	for k, v := range a.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}
	return labelSet
}
