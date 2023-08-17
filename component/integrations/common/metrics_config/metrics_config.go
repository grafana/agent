package common

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	internal "github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/prometheus/prometheus/model/labels"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.v2.common.metrics_config",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	autoscrapeConfig autoscrape.Config `river:"autoscrape,attr,optional`
	instanceKey      string            `river:"instance_key,string,optional"`
	extraLabels      map[string]string `river:"extra_labels,attr,optional`
}

type Exports struct {
	MetricsConfig internal.MetricsConfig `river:"metrics_config,attr"`
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

	mc := args.toInternalMetricsConfig()
	o.OnStateChange(Exports{MetricsConfig: mc})

	return c, nil
}

func (args *Arguments) toInternalMetricsConfig() internal.MetricsConfig {
	return internal.MetricsConfig{
		Autoscrape:  args.autoscrapeConfig,
		InstanceKey: &args.instanceKey,
		ExtraLabels: args.extraLabelsToInternals(),
	}
}

func (args *Arguments) extraLabelsToInternals() labels.Labels {
	var ls labels.Labels

	for name, value := range args.extraLabels {
		ls = append(ls, labels.Label{
			Name:  name,
			Value: value,
		})
	}
	return ls
}
