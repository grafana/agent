package app_agent_receiver

import (
	"context"
	"time"

	"github.com/grafana/agent/component"
	internal "github.com/grafana/agent/pkg/integrations/v2/app_agent_receiver"
	"github.com/grafana/agent/pkg/integrations/v2/common"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.v2.app_agent_receiver",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	metricsConfig   common.MetricsConfig `river:"metrics_config,attr,optional"`
	serverConfig    serverConfig         `river:"server_config,block,optional"`
	tracesInstance  string               `river:"traces_instance,string,optional"`
	logsInstance    string               `river:"logs_instance,string,optional"`
	logsLabels      map[string]string    `river:"log_labels,block,optional"`
	logsSendTimeout time.Duration        `river:"logs_send_time_out,attr,optional"`
	sourceMaps      sourceMapConfig      `river:"source_maps,block,optional"`
}

type Exports struct {
	Config internal.Config `river:"self,attr"`
}

type Component struct{}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{}

	o.OnStateChange(Exports{Config: args.toInternalConfig()})

	return c, nil
}

func (args *Arguments) toInternalConfig() internal.Config {
	return internal.Config{
		Common:          args.metricsConfig,
		Server:          args.serverConfig.toInternal(),
		TracesInstance:  args.tracesInstance,
		LogsInstance:    args.logsInstance,
		LogsLabels:      args.logsLabels,
		LogsSendTimeout: args.logsSendTimeout,
		SourceMaps:      args.sourceMaps.toInternal(),
	}
}
