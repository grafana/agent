package http

import (
	"context"
	"github.com/efficientgo/core/errors"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/http/internal/lokipush"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
	"sync"
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
	opts        component.Options
	entriesChan chan loki.Entry
	lock        sync.RWMutex
	args        Arguments    // guarded by lock
	cleanUp     func() error // guarded by lock
}

func init() {
	component.Register(component.Registration{
		Name: componentName,
		Args: Arguments{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

func New(opts component.Options, args Arguments) (component.Component, error) {
	c := &Component{
		opts:        opts,
		args:        args,
		entriesChan: make(chan loki.Entry),
		cleanUp:     func() error { return nil },
	}
	err := c.Update(args)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Component) Run(ctx context.Context) (err error) {
	defer func() {
		err = c.cleanUp()
	}()

	for {
		select {
		case entry := <-c.entriesChan:
			for _, receiver := range c.args.ForwardTo {
				receiver <- entry
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	newArgs, ok := args.(Arguments)
	if !ok {
		return errors.Newf("invalid type of arguments: %T", args)
	}

	pushTargetConfig := &lokipush.PushTargetConfig{
		Server: server.Config{
			HTTPListenPort:          newArgs.HTTPPort,
			HTTPListenAddress:       newArgs.HTTPAddress,
			Registerer:              c.opts.Registerer,
			MetricsNamespace:        "loki_source_http",
			RegisterInstrumentation: false,
			Log:                     logging.GoKit(c.opts.Logger),
		},
		Labels:        newArgs.labelSet(),
		KeepTimestamp: newArgs.UseIncomingTimestamp,
	}
	pushTarget, err := lokipush.NewPushTarget(
		c.opts.Logger,
		// When PushTarget is stopped, it will also Stop() the entry handler.
		loki.NewEntryHandler(c.entriesChan, func() {}),
		relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules),
		c.opts.ID,
		pushTargetConfig,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create loki push API server: %v", err)
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.args = newArgs
	c.cleanUp = func() error {
		return pushTarget.Stop()
	}
	return nil
}

func (a *Arguments) labelSet() model.LabelSet {
	labelSet := make(model.LabelSet, len(a.Labels))
	for k, v := range a.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}
	return labelSet
}
