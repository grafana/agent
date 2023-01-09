package relabel

import (
	"context"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.relabel",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the discovery.relabel component.
type Arguments struct {
	// Targets contains the input 'targets' passed by a service discovery component.
	Targets []discovery.Target `river:"targets,attr"`

	// The relabelling rules to apply to the each target's label set.
	RelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`
}

// Exports holds values which are exported by the discovery.relabel component.
type Exports struct {
	Output []discovery.Target `river:"output,attr"`
	Rules  flow_relabel.Rules `river:"rules,attr"`
}

// Component implements the discovery.relabel component.
type Component struct {
	opts component.Options
	rcs  []*relabel.Config
}

var _ component.Component = (*Component)(nil)

// New creates a new discovery.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	targets := make([]discovery.Target, 0, len(newArgs.Targets))
	relabelConfigs := flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelConfigs)
	c.rcs = relabelConfigs

	for _, t := range newArgs.Targets {
		lset := componentMapToPromLabels(t)
		lset = relabel.Process(lset, relabelConfigs...)
		if lset != nil {
			targets = append(targets, promLabelsToComponent(lset))
		}
	}

	c.opts.OnStateChange(Exports{
		Output: targets,
		Rules:  c.getRules,
	})

	return nil
}

func (c *Component) getRules() []*relabel.Config {
	return c.rcs
}

func componentMapToPromLabels(ls discovery.Target) labels.Labels {
	res := make([]labels.Label, 0, len(ls))
	for k, v := range ls {
		res = append(res, labels.Label{Name: k, Value: v})
	}

	return res
}

func promLabelsToComponent(ls labels.Labels) discovery.Target {
	res := make(map[string]string, len(ls))
	for _, l := range ls {
		res[l.Name] = l.Value
	}

	return res
}
