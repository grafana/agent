package http

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/grafana/agent/pkg/util"

	"github.com/go-kit/log/level"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/http/internal/lokipush"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.http",
		Args: Arguments{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	HTTPAddress          string              `river:"http_address,attr"`
	HTTPPort             int                 `river:"http_port,attr"`
	ForwardTo            []loki.LogsReceiver `river:"forward_to,attr"`
	Labels               map[string]string   `river:"labels,attr,optional"`
	RelabelRules         relabel.Rules       `river:"relabel_rules,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	// TODO: allow to configure other Server fields in a dedicated block, to match promtail's
	//       https://grafana.com/docs/loki/next/clients/promtail/configuration/#server
}

func (a *Arguments) labelSet() model.LabelSet {
	labelSet := make(model.LabelSet, len(a.Labels))
	for k, v := range a.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}
	return labelSet
}

type Component struct {
	opts         component.Options
	entriesChan  chan loki.Entry
	unregisterer *util.Unregisterer

	rwMut      sync.RWMutex
	args       Arguments
	pushTarget *lokipush.PushTarget

	// Use separate receivers mutex to address potential deadlock when Update drains the current server.
	// e.g. https://github.com/grafana/agent/issues/3391
	receiversMut sync.RWMutex
	receivers    []loki.LogsReceiver
}

func New(opts component.Options, args Arguments) (component.Component, error) {
	c := &Component{
		opts:         opts,
		args:         args,
		entriesChan:  make(chan loki.Entry),
		receivers:    args.ForwardTo,
		unregisterer: util.WrapWithUnregisterer(opts.Registerer),
	}
	err := c.Update(args)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Component) Run(ctx context.Context) (err error) {
	defer func() {
		err = c.stop()
	}()

	for {
		select {
		case entry := <-c.entriesChan:
			c.receiversMut.RLock()
			receivers := c.receivers
			c.receiversMut.RUnlock()

			for _, receiver := range receivers {
				select {
				case receiver <- entry:
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

	c.receiversMut.Lock()
	c.receivers = newArgs.ForwardTo
	c.receiversMut.Unlock()

	newPushTargetConfig := c.pushTargetConfigForArgs(newArgs)

	c.rwMut.Lock()
	defer c.rwMut.Unlock()

	pushTargetNeedsUpdate := c.pushTarget == nil || !reflect.DeepEqual(c.pushTarget.CurrentConfig(), *newPushTargetConfig)
	if !pushTargetNeedsUpdate {
		c.args = newArgs
		return nil
	}

	if c.pushTarget != nil {
		err := c.pushTarget.Stop()
		if err != nil {
			level.Warn(c.opts.Logger).Log("msg", "push API server failed to stop on update", "err", err)
		}
		c.pushTarget = nil
		c.unregisterer.UnregisterAll()
	}

	newPushTarget, err := lokipush.NewPushTarget(
		c.opts.Logger,
		loki.NewEntryHandler(c.entriesChan, func() {}),
		c.opts.ID,
		newPushTargetConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to create loki push API server: %v", err)
	}

	c.pushTarget = newPushTarget
	c.args = newArgs
	return nil
}

func (c *Component) pushTargetConfigForArgs(newArgs Arguments) *lokipush.PushTargetConfig {
	return &lokipush.PushTargetConfig{
		Server: server.Config{
			HTTPListenPort:          newArgs.HTTPPort,
			HTTPListenAddress:       newArgs.HTTPAddress,
			Registerer:              c.unregisterer,
			MetricsNamespace:        "loki_source_http",
			RegisterInstrumentation: false,
			Log:                     logging.GoKit(c.opts.Logger),
		},
		Labels:        newArgs.labelSet(),
		KeepTimestamp: newArgs.UseIncomingTimestamp,
		RelabelConfig: relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules),
	}
}

func (c *Component) stop() error {
	c.rwMut.RLock()
	defer c.rwMut.RUnlock()
	if c.pushTarget != nil {
		err := c.pushTarget.Stop()
		c.unregisterer.UnregisterAll()
		return err
	}
	return nil
}
