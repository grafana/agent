package http

import (
	"context"
	"reflect"
	"sync"

	"github.com/grafana/agent/pkg/util"

	"github.com/go-kit/log/level"

	"github.com/efficientgo/core/errors"
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
	rwLock       sync.RWMutex

	// The following fields must be guarded by the rwLock
	args       Arguments
	pushTarget *lokipush.PushTarget
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
		opts:         opts,
		args:         args,
		entriesChan:  make(chan loki.Entry),
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
			c.rwLock.RLock()
			forwardTo := c.args.ForwardTo
			c.rwLock.RUnlock()

			for _, receiver := range forwardTo {
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
		return errors.Newf("invalid type of arguments: %T", args)
	}

	newPushTargetConfig := c.pushTargetConfigForArgs(newArgs)

	c.rwLock.Lock()
	defer c.rwLock.Unlock()

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
		return errors.Wrapf(err, "failed to create loki push API server: %v", err)
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
	c.rwLock.RLock()
	defer c.rwLock.RUnlock()
	if c.pushTarget != nil {
		err := c.pushTarget.Stop()
		c.unregisterer.UnregisterAll()
		return err
	}
	return nil
}
