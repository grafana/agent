package mutate

import (
	"context"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name:    "targets.mutate",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the targets.mutate component.
type Arguments struct {
	// Targets contains the input 'targets' passed by a service discovery component.
	Targets []Target `river:"targets,attr"`

	// The relabelling steps to apply to the each target's label set.
	RelabelConfigs []*flow_relabel.Config `river:"relabel_config,block,optional"`
}

// Target refers to a singular HTTP or HTTPS endpoint that will be used for scraping.
// Here, we're using a map[string]string instead of labels.Labels; if the label ordering
// is important, we can change to follow the upstream logic instead.
// TODO (@tpaschalis) Maybe the target definitions should be part of the
// Service Discovery components package. Let's reconsider once it's ready.
type Target map[string]string

// Exports holds values which are exported by the targets.mutate component.
type Exports struct {
	Output []Target `river:"output,attr"`
}

// Component implements the targets.mutate component.
type Component struct {
	opts component.Options
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new targets.mutate component.
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

	targets := make([]Target, 0, len(newArgs.Targets))
	relabelConfigs := flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelConfigs)

	for _, t := range newArgs.Targets {
		lset := componentMapToPromLabels(t)
		lset = relabel.Process(lset, relabelConfigs...)
		if lset != nil {
			targets = append(targets, promLabelsToComponent(lset))
		}
	}

	c.opts.OnStateChange(Exports{
		Output: targets,
	})

	return nil
}

func componentMapToPromLabels(ls Target) labels.Labels {
	res := make([]labels.Label, 0, len(ls))
	for k, v := range ls {
		res = append(res, labels.Label{Name: k, Value: v})
	}

	return res
}

func promLabelsToComponent(ls labels.Labels) Target {
	res := make(map[string]string, len(ls))
	for _, l := range ls {
		res[l.Name] = l.Value
	}

	return res
}
